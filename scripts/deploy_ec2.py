from __future__ import annotations

import argparse
import base64
import hashlib
import io
import json
import logging
import os
import queue
import subprocess
import threading
import time
import uuid
from pathlib import Path, PurePosixPath
from typing import Any, Callable, Dict, List, Optional

import boto3
import boto3.session
import paramiko

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s.%(msecs)03d  %(name)-12s %(levelname)-8s %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)

logger = logging.getLogger(__name__)


# =================== EC2 Management ====================
class EC2Config():
    def __init__(self, image_id: str, instance_type: str) -> None:
        self.instance_type = instance_type
        self.block_device_mappings = []
        self.image_id = image_id
        self.ip_permissions = []
        self.user_data = ''

    @staticmethod
    def get_latest_ubuntu_ami() -> str:
        session = boto3.session.Session()
        images = session.client('ec2').describe_images(Filters=[{
            'Name': 'name',
            'Values': ['ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server*']
        }])
        return sorted(images['Images'], key=lambda i: i['CreationDate'])[-1]['ImageId']

    def with_storage(
            self, volume_size: int = 8,
            volume_type: str = 'gp2',
            device_name: str = '/dev/sda1',
            delete_on_termination: bool = True) -> EC2Config:
        self.block_device_mappings.append({
            'DeviceName': device_name,
            'Ebs': {
                'DeleteOnTermination': delete_on_termination,
                'VolumeSize': volume_size,
                'VolumeType': volume_type
            },
        })
        return self

    def with_inbound_rule(self, protocol: str, port: int, ip_range: str = '0.0.0.0/0') -> EC2Config:
        self.ip_permissions.append({
            'IpProtocol': protocol,
            'FromPort': port,
            'ToPort': port,
            'IpRanges': [{'CidrIp': ip_range}]
        })
        return self

    def with_user_data(self, user_data: str) -> EC2Config:
        self.user_data = user_data
        return self


class EC2RuntimeException(Exception):
    pass


class EC2KeyPair():
    DEFAULT_USER = 'ubuntu'
    DEFAULT_PORT = 22

    def __init__(self, key_name: str, priv_key: paramiko.PKey) -> None:
        self.key_name = key_name
        self.priv_key = priv_key

    @staticmethod
    def from_file(filename: str) -> EC2KeyPair:
        f = argparse.FileType('r')(filename)
        name = os.path.basename(f.name).replace('.pem', '')
        key = paramiko.RSAKey.from_private_key(f)
        return EC2KeyPair(key_name=name, priv_key=key)

    @staticmethod
    def from_str(keystr: str) -> EC2KeyPair:
        name, key_str = keystr.split(':')
        key = paramiko.RSAKey.from_private_key(io.StringIO(key_str))
        return EC2KeyPair(key_name=name, priv_key=key)

    @staticmethod
    def new(filename: str) -> EC2KeyPair:
        f = argparse.FileType('x')(filename)
        name = os.path.basename(f.name).replace('.pem', '')
        logger.info(f'Creating a new EC2 key pair {name}...')
        keypair = boto3.client('ec2').create_key_pair(KeyName=name)
        f.writelines(keypair['KeyMaterial'])
        f.close()
        logger.info(f'Saved EC2 key pair to {filename}')
        key = paramiko.RSAKey.from_private_key_file(filename)
        return EC2KeyPair(key_name=name, priv_key=key)


class EC2SSHConfig():
    def __init__(self, keypair: EC2KeyPair, username: str = 'ubuntu', port: int = 22) -> None:
        self.keypair = keypair
        self.username = username
        self.port = port


class EC2Instance():
    NotExistsException = EC2RuntimeException('Instance does not exist')
    AlreadyExistsException = EC2RuntimeException('Instance already exists')

    def __init__(self, app_name: str, ssh_config: EC2SSHConfig,
                 logger: logging.Logger = None) -> None:
        self.app_name = app_name
        self.ssh_config = ssh_config

        self._boto3_session = boto3.session.Session()
        self._ec2_client = self._boto3_session.client('ec2')
        self._session_id = str(uuid.uuid4())
        self._logger = logger or logging.getLogger(self.__class__.__name__)

        self._ssh = None

    @property
    def instance(self) -> Optional[Dict[str, Any]]:
        response = self._ec2_client.describe_instances(Filters=[{
            'Name': 'tag:app_name',
            'Values': [self.app_name]
        }])
        instances = []
        for reservation in response['Reservations']:
            for instance in reservation['Instances']:
                if instance['State']['Name'] == 'running':
                    instances.append(instance)
        if len(instances) > 1:
            raise EC2RuntimeException(
                f'Found multiple instance with tag app_name:{self.app_name}! '
                'Please manually check them in AWS console.')
        if len(instances) == 1:
            return instances[0]

    @property
    def exists(self) -> bool:
        return self.instance is not None

    @property
    def public_ip(self) -> str:
        if not self.exists:
            raise self.NotExistsException
        return self.instance['PublicIpAddress']

    @property
    def private_ip(self) -> str:
        if not self.exists:
            raise self.NotExistsException
        return self.instance['PrivateIpAddress']

    @property
    def instace_id(self) -> str:
        if not self.exists:
            raise self.NotExistsException
        return self.instance['InstanceId']

    @property
    def ssh(self) -> paramiko.SSHClient:
        if self._ssh:
            return self._ssh

        if not self.exists:
            raise self.NotExistsException

        retries = 5

        self._ssh = paramiko.SSHClient()
        self._ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        ip_address = self.public_ip
        for try_count in range(retries):
            try:
                self._logger.info(f'SSH into the instance {self.app_name}({ip_address})')
                self._ssh.connect(
                    ip_address,
                    port=self.ssh_config.port,
                    username=self.ssh_config.username,
                    pkey=self.ssh_config.keypair.priv_key)
                return self._ssh
            except Exception as e:
                interval = try_count * 5 + 5
                self._logger.warning(e)
                self._logger.info(f'Retrying in {interval} seconds...')
                time.sleep(interval)
        raise EC2RuntimeException(f'Failed to ssh into the {self.app_name}({ip_address})')

    def _create_security_group(self, ip_permissions):
        rule_id = hashlib.md5(str(ip_permissions).encode('utf-8')).hexdigest()[:8]
        group_name = self.app_name + '-' + rule_id

        # return existing group if exists
        response = self._ec2_client.describe_security_groups(Filters=[{
            'Name': 'group-name',
            'Values': [group_name]}
        ])
        if len(response['SecurityGroups']) > 0:
            return response['SecurityGroups'][0]['GroupId']

        # Create new security group
        response = self._ec2_client.create_security_group(
            GroupName=group_name, Description=group_name)
        security_group_id = response['GroupId']
        self._ec2_client.authorize_security_group_ingress(
            GroupId=security_group_id,
            IpPermissions=ip_permissions
        )
        # There might be some consistency issue without sleep
        time.sleep(5)
        return security_group_id

    def wait_for_cloud_init(self) -> None:
        self.run_command('; '.join([
            "while [ ! -f /var/lib/cloud/instance/boot-finished ]",
            "do echo 'Waiting for cloud-init...'",
            "sleep 3",
            "done"
        ]))

    def launch(self, config: EC2Config, wait_init: bool = True) -> None:
        if self.exists:
            raise self.AlreadyExistsException

        self._ec2_client.run_instances(
            ImageId=config.image_id,
            InstanceType=config.instance_type,
            MinCount=1,
            MaxCount=1,
            BlockDeviceMappings=config.block_device_mappings,
            SecurityGroupIds=[
                self._create_security_group(config.ip_permissions)
            ],
            TagSpecifications=[{
                'ResourceType': 'instance',
                'Tags': [
                    {
                        'Key': 'app_name',
                        'Value': self.app_name
                    },
                    {
                        'Key': 'Name',
                        'Value': self.app_name
                    }
                ]
            }],
            KeyName=self.ssh_config.keypair.key_name,
            UserData=config.user_data
        )

        while self.instance is None:
            self._logger.info('Waitting for the instance to be running...')
            time.sleep(10)

        self._logger.info('Waitting for ssh to be ready(about 30s)...')
        time.sleep(30)

        if wait_init:
            self.wait_for_cloud_init()

    def terminate(self) -> None:
        if not self.exists:
            raise self.NotExistsException
        ec2 = self._boto3_session.resource('ec2')
        ec2.instances.filter(InstanceIds=[self.instace_id]).terminate()

    def run_command(self, command: str) -> List[str]:
        def channel_logger(logger_func, channel):
            while True:
                line = channel.readline()
                if not line:
                    break
                logger_func(line.rstrip())

        self._logger.info('$ ' + command)
        session_file = '/tmp/ssh_seesion_' + self._session_id

        for _ in range(3):
            try:
                _, stdout, stderr = self.ssh.exec_command(' ; '.join([
                    'touch ' + session_file,
                    'source ' + session_file,
                    'cd $PWD',
                    command + ' && export -p > ' + session_file
                ]))
                break
            except ConnectionResetError as e:
                self._logger.warning(e)
                self._ssh = None  # reset ssh
        else:
            raise EC2RuntimeException(
                'An existing connection was forcibly closed by the remote host')

        stdout_lines = []

        t_err = threading.Thread(target=channel_logger, args=(self._logger.warning, stderr))
        t_err.start()
        try:
            channel_logger(lambda line: (
                self._logger.info(line),
                stdout_lines.append(line)
            ), stdout)
            t_err.join()
            if stdout.channel.recv_exit_status() != 0:
                raise EC2RuntimeException(
                    f'Command `{command}` failed with non-zero exit code')
        except (KeyboardInterrupt, Exception):
            stderr.channel.close()
            self._logger.info('Channel closed')
            raise

        return stdout_lines

    def run_local_script(
            self, script_path: Path,
            sudo: bool = False, cwd: Optional[PurePosixPath] = None) -> None:
        remote_path = '/tmp/automation_script'
        self.upload_file(script_path, PurePosixPath(remote_path))
        # this line converts CRLF to LF in case Windows in used to deploy
        self.run_command('ex -bsc \'%!awk "{sub(/\\r/,\\"\\")}1"\' -cx ' + remote_path)
        self.run_command('chmod +x ' + remote_path)
        commands = []
        if cwd:
            commands.append('cd ' + str(cwd))
        commands.append(('sudo -E ' if sudo else '') + '/tmp/automation_script')
        self.run_command(' && '.join(commands))

    def upload_file(self, local_path: Path, remote_path: PurePosixPath) -> None:
        try:
            with self.ssh.open_sftp() as sftp:
                self._logger.info(f'Uploading {str(local_path.absolute())} -> {str(remote_path)}')
                sftp.put(localpath=str(local_path.absolute()), remotepath=str(remote_path))
        except ConnectionResetError as e:
            self._logger.warning(e)
            self._ssh = None
            self.upload_file(local_path, remote_path)

    def download_file(self, remote_path: PurePosixPath, local_path: Path) -> Path:
        try:
            with self.ssh.open_sftp() as sftp:
                sftp.get(remotepath=str(remote_path), localpath=str(local_path.absolute()))
                return local_path
        except ConnectionResetError as e:
            self._logger.warning(e)
            self._ssh = None
            return self.download_file(remote_path, local_path)

    def import_variable(self, **kvs: str):
        for name, value in kvs.items():
            encoded = base64.b64encode(value.encode('utf-8')).decode('ascii')
            self.run_command(f'export {name}=$(echo {encoded} | base64 -d)')

    def export_variable(self, name: str) -> str:
        encoded = ''.join(self.run_command(f'echo -n "${name}" | base64'))
        value = base64.decodebytes(encoded.encode('ascii')).decode('utf-8')
        return value


# =================== Helpers ====================
def archive_git_dir(dir: Path, save_as: Path) -> Path:
    p = subprocess.Popen(['git', 'archive', '-o', str(save_as), 'HEAD'], cwd=dir)
    p.wait()
    if p.returncode != 0:
        raise RuntimeError(f'Failed to create project archive: {p.stderr}')
    return save_as


class FutureValue():
    def __init__(self) -> None:
        self._queue = queue.Queue(maxsize=1)
        self._done = False
        self._result = None
        self._result_saved = False
        self._w_lock = threading.Lock()
        self._r_lock = threading.Lock()

    def set(self, value) -> None:
        with self._w_lock:
            if self._done:
                raise ValueError('Future can only be set once')
            self._queue.put_nowait(value)
            self._done = True

    def get(self) -> Any:
        with self._r_lock:
            if self._result_saved:
                return self._result
            self._result = self._queue.get()
            self._result_saved = True
            return self._result


class CatchableThread(threading.Thread):
    def __init__(self, target: Callable[[], Any]) -> None:
        super().__init__()
        self._target = target
        self._exception = None
        self._lock = threading.Lock()

    def run(self):
        try:
            self._target()
        except BaseException as e:
            self._lock.acquire()
            self._exception = e
            self._lock.release()
            raise

    def exception(self):
        self._lock.acquire()
        e = self._exception
        self._lock.release()
        return e


def run_in_parallel(*tasks: Callable[[], Any]) -> None:
    threads = [CatchableThread(target=task) for task in tasks]
    for thread in threads:
        thread.daemon = True
        thread.start()
    pending = threads[:]
    while pending:
        for i in range(len(pending)):
            # join thread with timeout 0.5s
            pending[i].join(timeout=0.5)
            # if the thread is still alive, means we reached timeout
            if pending[i].is_alive():
                # just continue
                continue
            # otherwise the thread did finished
            # if due to exception, abort all tasks
            e = pending[i].exception()
            if e:
                raise Exception('Abort all tasks because of exception: ' + str(e))
            # remove the finished task from pending
            pending.pop(i)
            break


# ================= Main Function =================
def launch(ssh_config: EC2SSHConfig, num_nodes: int, num_replicas: int):
    project_base = Path(__file__).absolute().parent.parent
    project_archive = archive_git_dir(project_base, save_as=project_base/'archive.tar.gz')

    node_private_ips = [FutureValue() for _ in range(num_nodes+1)]
    node_public_ips = [FutureValue() for _ in range(num_nodes+1)]

    def get_cluster_config():
        num_read = num_replicas // 2 + 1
        num_write = num_read
        return {
            "NumReplica": num_replicas,
            "NumRead": num_read,
            "NumWrite": num_write,
            "NumVirtualNodes": 10,
            "Timeout": 500,
            "SeedServers": [{
                "Id": node_id,
                "IpInternal": node_private_ips[node_id].get(),
                "IpExternal": node_private_ips[node_id].get(),
                "PortInternal": 5000 + node_id,
                "PortExternal": 6000 + node_id,
            } for node_id in range(1, num_nodes+1)]
        }

    def get_client_config():
        return {
            "NodeTimeoutMs": 1000,
            "Retry": 4,
            "SeedNodes": [{
                "ID": 5000 + node_id,
                "Address": f"{node_private_ips[node_id].get()}:{6000+node_id}"
            } for node_id in range(1, num_nodes+1)]
        }

    def setup_go_env(logger, instance):
        logger.info('Uploading project archive')
        instance.upload_file(project_archive, PurePosixPath('app.tar.gz'))

        logger.info('Releasing files')
        instance.run_command(' && '.join([
            'mkdir -p app',
            'tar -zxf app.tar.gz -C app'
        ]))

        logger.info('Installing go')
        instance.run_command(' && '.join([
            'wget -q -c https://golang.org/dl/go1.16.3.linux-amd64.tar.gz',
            'sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.16.3.linux-amd64.tar.gz',
            'echo "export PATH=$PATH:/usr/local/go/bin" >> .bashrc',
            'export PATH=$PATH:/usr/local/go/bin',
            'go version'
        ]))

        logger.info('Initializing go module')
        instance.run_command(' && '.join([
            'cd app',
            'go mod vendor'
        ]))

    def deploy_node(node_id):
        def do_deploy_node():
            logger = logging.getLogger(f'@node{node_id}')
            instance = EC2Instance(f'node{node_id}', ssh_config, logger)

            if not instance.exists:
                if instance.exists:
                    logger.info('Terminating existing node instance')
                    instance.terminate()

                logger.info('Launching new node instance')
                instance.launch(
                    EC2Config(
                        image_id=EC2Config.get_latest_ubuntu_ami(),
                        instance_type='t2.micro')
                    .with_storage(volume_size=8)
                    .with_inbound_rule('tcp', 22)
                    .with_inbound_rule('tcp', 5000+node_id, '172.31.0.0/16')
                    .with_inbound_rule('tcp', 6000+node_id, '172.31.0.0/16'))

            node_private_ips[node_id].set(instance.private_ip)
            node_public_ips[node_id].set(instance.public_ip)
            setup_go_env(logger, instance)

            logger.info('Starting server')
            config = get_cluster_config()
            instance.import_variable(CLUSTER_CONFIG=json.dumps(config))
            instance.run_command('echo $CLUSTER_CONFIG > config.json')
            instance.import_variable(
                LAUNCH_SCRIPT='#!/bin/sh\n'
                + 'nohup '
                + f'go run ./cmd/server -seed -index {node_id} -config config.json '
                + '> ./log.txt 2>&1 &')
            instance.run_command(' && '.join([
                'echo "$LAUNCH_SCRIPT" > launch',
                'chmod +x launch',
                './launch'
            ]))

        return do_deploy_node

    client_public_ip = FutureValue()

    def deploy_client():
        logger = logging.getLogger('@client')
        instance = EC2Instance('client', ssh_config, logger)

        if not instance.exists:
            if instance.exists:
                logger.info('Terminating existing node instance')
                instance.terminate()

            logger.info('Launching new node instance')
            instance.launch(
                EC2Config(
                    image_id=EC2Config.get_latest_ubuntu_ami(),
                    instance_type='t2.micro')
                .with_storage(volume_size=8)
                .with_inbound_rule('tcp', 22)
                .with_inbound_rule('tcp', 8080))

        client_public_ip.set(instance.public_ip)
        setup_go_env(logger, instance)

        logger.info('Starting server')
        config = get_client_config()
        instance.import_variable(CLIENT_CONFIG=json.dumps(config))
        instance.run_command('echo $CLIENT_CONFIG > config.json')
        instance.import_variable(
            LAUNCH_SCRIPT='#!/bin/sh\n'
            + 'nohup '
            + 'go run ./cmd/safeentry -address :8080 -config config.json '
            + '> ./log.txt 2>&1 &')
        instance.run_command(' && '.join([
            'echo "$LAUNCH_SCRIPT" > launch',
            'chmod +x launch',
            './launch'
        ]))

    # Run script
    tasks = [deploy_client]
    for node_id in range(1, 1+num_nodes):
        tasks.append(deploy_node(node_id))
    run_in_parallel(*tasks)

    print('\n\x1b[6;30;42m' + 'Cluster is Ready!' + '\x1b[0m')
    print()

    for node_id in range(1, 1+num_nodes):
        print(f'\x1b[1;32;40m●\x1b[0m Node {node_id}  {node_public_ips[node_id].get()}')
    print()

    print('\x1b[1;32;40m●\x1b[0m Safe Entry Server')
    print('IP: ', client_public_ip.get())
    print('URL:', f'http://{client_public_ip.get()}:8080/')
    print()

    # Clean up
    os.remove(project_archive)


def terminate(ssh_config: EC2SSHConfig):
    logger = logging.getLogger('terminate')
    logger.info('Terminating cluster')

    # Terminate data nodes
    not_found_count = 0
    data_node_id = 0
    while not_found_count < 3:
        data_node = EC2Instance(f'node{data_node_id}', ssh_config, logger)
        if data_node.exists:
            logger.info(f'Terminating data node {data_node_id}')
            data_node.terminate()
            not_found_count = 0
        else:
            not_found_count += 1
        data_node_id += 1

    # Terminate client
    client_node = EC2Instance('client', ssh_config, logger)
    if client_node.exists:
        logger.info('Terminating client node')
        client_node.terminate()


def main():
    parser = argparse.ArgumentParser()
    keypair_group = parser.add_mutually_exclusive_group(required=True)
    keypair_group.add_argument('--keyfile', type=EC2KeyPair.from_file)
    keypair_group.add_argument('--key', type=EC2KeyPair.from_str)
    keypair_group.add_argument('--newkey', type=EC2KeyPair.new, dest='new_key_path')
    subparsers = parser.add_subparsers(dest='action')
    launch_parser = subparsers.add_parser('launch')
    launch_parser.add_argument('--nodes', type=int, required=True)
    launch_parser.add_argument('--replicas', type=int, required=True)
    subparsers.add_parser('terminate')
    args = parser.parse_args()

    keypair = args.keyfile or args.key or args.new_key_path
    ssh_config = EC2SSHConfig(keypair, username='ubuntu', port=22)

    if args.action is None or args.action == 'launch':
        launch(ssh_config, args.nodes, args.replicas)
    elif args.action == 'terminate':
        terminate(ssh_config)


if __name__ == "__main__":
    main()
