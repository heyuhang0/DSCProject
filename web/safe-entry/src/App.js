import { useState } from 'react';
import { Card, Form, Input, Button, Switch, List } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import './App.css';


function LocationList({ setLocation }) {
  const [lastLocation, setLastLocation] = useState('SUTD Building 1');
  // setLastLocation('SUTD Building 1');

  const data = [
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
          <h3>SUTD</h3>
          <Button>Check out</Button>
        </Card>: null
        }
      </div>
      
      <h3>Where are you visiting?</h3>
      <List
        dataSource={data}
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
