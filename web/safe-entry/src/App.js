import { useState } from 'react';
import { Card, Form, Input, Button, Switch, List } from 'antd';
import { ArrowLeftOutlined, ExportOutlined, ImportOutlined } from '@ant-design/icons';
import './App.css';


function LocationList({ setLocation }) {
  // setLastLocation('SUTD Building 1');

  const dummyHistory = [
    {
      location: 'SUTD Block 55',
      action: 'check-in'
    },
    {
      location: 'SUTD Block 55',
      action: 'check-out'
    },
    {
      location: 'SUTD Building 2',
      action: 'check-in'
    },
    {
      location: 'SUTD Building 2',
      action: 'check-out'
    },
    {
      location: 'SUTD Building 1',
      action: 'check-in'
    }
  ];

  const [history, setHistory] = useState(dummyHistory);

  // find the last check in location that has not been checked out
  let checkInRecords = history.filter(record => record.action === 'check-in');
  let checkOutRecords = history.filter(record => record.action !== 'check-in');
  checkOutRecords.forEach(outRecord => {
    for (let i = 0; i < checkInRecords.length; i ++) {
      if (checkInRecords[i].location === outRecord.location) {
        checkInRecords = checkInRecords.filter((_, index) => index !== i);
        break;
      }
    }
  })
  let lastLocation = null;
  if (checkInRecords.length > 0) {
    lastLocation = checkInRecords[checkInRecords.length - 1].location;
  }

  const locations = [
    'SUTD Building 1',
    'SUTD Building 2',
    'SUTD Building 3',
    'SUTD Building 5',
    'SUTD Block 55',
    'SUTD Block 57',
    'SUTD Block 59',
    'SUTD Block 61',
  ];
  return (
    <div>
      <div>
        {lastLocation !== null ?
          <Card title='Last check-in' style={{ width: 300, margin: '40px auto' }}>
            <h3>{lastLocation}</h3>
            <Button>Check out</Button>
          </Card> : null
        }
      </div>

      <h3>Where are you visiting?</h3>
      <List
        dataSource={locations}
        split={false}
        renderItem={item => (
          <List.Item>
            <Button
              style={{ margin: "auto", width: 300 }}
              onClick={() => setLocation(item)}
            >
              {item}
            </Button>
          </List.Item>
        )}
      />

      <Card title="History" style={{ width: 300, margin: '40px auto' }}>
        <List
          dataSource={history}
          renderItem={record => (
            <List.Item>
              <span>
                {record.action === 'check-in' ?
                  <ImportOutlined style={{color: "green"}} /> :
                  <ExportOutlined style={{color: "red"}} />}
              </span>
              <span>{record.location}</span>
            </List.Item>
          )}
        />
      </Card>
    </div>
  );
}


function CheckInPage({ location }) {
  let action = "check-in";

  let onFinish = (form) => {
    console.log(action, form);
  };

  return (
    <div>
      <Card title={location} style={{ width: 300, margin: '40px auto' }}>
        <Form
          name="basic"
          initialValues={{ remember: true }}
          onFinish={onFinish}
        >
          <Form.Item
            name="ic"
            rules={[{ required: true, message: 'Please input your NRIC/FIN' }]}
          >
            <Input placeholder="NRIC/FIN:" />
          </Form.Item>

          <Form.Item
            name="mobile"
            rules={[{ required: true, message: 'Please input your mobile number' }]}
          >
            <Input placeholder="Mobile Number:" />
          </Form.Item>

          <Form.Item name="remember" valuePropName="checked" label="Remember my particulars:">
            <Switch />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              size="large"
              shape="round"
              style={{ background: "#1FAF71", borderColor: "#1FAF71", width: 180 }}
              onClick={() => { action = "check-in" }}
            >
              Check-In
            </Button>
          </Form.Item>

          <Form.Item>
            <Button
              htmlType="submit"
              size="large"
              shape="round"
              style={{ width: 180 }}
              onClick={() => { action = "check-out" }}
            >
              Check-Out
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}

function App() {
  const [location, setLocation] = useState(null);

  return (
    <div className="App">
      <div>
        {location !== null ?
          <Button
            type="text"
            style={{
              position: 'absolute',
              left: '50%',
              marginLeft: -160,
              top: 20
            }}
            onClick={() => setLocation(null)}
          >
            <ArrowLeftOutlined />Back
          </Button> :
          null
        }
      </div>

      <h1 style={{ marginTop: 20 }}>Safe<b>Entry</b></h1>
      <div>
        {location === null ?
          <LocationList setLocation={setLocation} /> :
          <CheckInPage location={location} />
        }
      </div>

    </div>
  );
}

export default App;
