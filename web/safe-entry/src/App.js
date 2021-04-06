import { useState, useEffect } from 'react';
import { Card, Form, Input, Button, Switch, List } from 'antd';
import { ArrowLeftOutlined, ExportOutlined, ImportOutlined } from '@ant-design/icons';
import axios from 'axios';
import './App.css';


function LocationList({ setLocation }) {
  const [history, setHistory] = useState([]);

  const refreshHistory = () => {
    let saved = localStorage.getItem('particulars');
    if (saved !== null) {
      saved = JSON.parse(saved);
      axios.post("/api/history", {
        "IC": saved.ic,
        "Phone": saved.phone,
      }).then(res => {
        setHistory(res.data.map(record => {
          return {
            location: record.Location,
            action: record.CheckIn ? 'check-in' : 'check-out'
          };
        }))
      });
    }
  }

  useEffect(refreshHistory, []);

  // find the last check in location that has not been checked out
  let checkInRecords = history.filter(record => record.action === 'check-in');
  let checkOutRecords = history.filter(record => record.action !== 'check-in');
  checkOutRecords.forEach(outRecord => {
    for (let i = 0; i < checkInRecords.length; i++) {
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
            <Button onClick={() => {
              let saved = localStorage.getItem('particulars');
              if (saved !== null) {
                saved = JSON.parse(saved);
                axios.post("/api/checkin", {
                  "IC": saved.ic,
                  "Phone": saved.phone,
                  "Location": lastLocation,
                  "CheckIn": false,
                }).then(refreshHistory);
              }
            }}>
              Check out
            </Button>
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
                  <ImportOutlined style={{ color: "green" }} /> :
                  <ExportOutlined style={{ color: "red" }} />}
              </span>
              <span>{record.location}</span>
            </List.Item>
          )}
        />
      </Card>
    </div>
  );
}


function CheckInPage({ location, afterSubmit }) {
  let action = "check-in";

  let onFinish = (form) => {
    if (form.remember) {
      localStorage.setItem("particulars", JSON.stringify(form));
    } else {
      localStorage.removeItem("particulars");
    }

    axios.post("/api/checkin", {
      "IC": form.ic,
      "Phone": form.phone,
      "Location": location,
      "CheckIn": action === "check-in",
    }).then(() => {
      afterSubmit();
    });
  };

  let saved = localStorage.getItem('particulars');
  let savedIC = null;
  let savedPhone = null;
  if (saved !== null) {
    saved = JSON.parse(saved);
    savedIC = saved.ic;
    savedPhone = saved.phone;
  }

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
            initialValue={savedIC}
          >
            <Input placeholder="NRIC/FIN:" />
          </Form.Item>

          <Form.Item
            name="phone"
            rules={[{ required: true, message: 'Please input your mobile number' }]}
            initialValue={savedPhone}
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
          <CheckInPage location={location} afterSubmit={() => setLocation(null)} />
        }
      </div>

    </div>
  );
}

export default App;
