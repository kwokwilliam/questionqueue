import React, { useState } from 'react';
import { Input, Button } from 'reactstrap';
import { Link } from 'react-router-dom';
import Endpoints from '../../../../../Endpoints/Endpoints';

export default function TutorQAdminMain({ adminButtons, signOut }) {
    const [classNum, setClassNum] = useState("");
    const [topics, setTopics] = useState("");
    const inputStyles = { maxWidth: 500, margin: 'auto', marginTop: 30 };

    return <>
        <div style={{ fontSize: '150%' }}>
            {adminButtons.map(d => {
                return <div key={d.linkTo} style={{ marginBottom: 5 }}>
                    <Link className="btn" style={{ textDecoration: 'none', color: 'white', backgroundColor: '#005696' }} to={d.linkTo}>{d.linkText}</Link>
                </div>
            })}
        </div>
        <hr />
        <div>
            <h2>Insert new class</h2>
            <Input placeholder={"Class number"}
                onChange={(e) => { setClassNum(e.target.value) }}
                value={classNum}
                style={inputStyles} />
            <Input placeholder={"topics"}
                onChange={(e) => { setTopics(e.target.value) }}
                value={topics}
                style={inputStyles} />
            <Button onClick={async () => {
                const { URL, ClassControl } = Endpoints;
                const sendData = {
                    class_number: classNum,
                    topics: topics.split(",")
                }
                const response = await fetch(URL + ClassControl, {
                    method: "POST",
                    body: JSON.stringify(sendData),
                    headers: new Headers({
                        "Content-Type": "application/json"
                    })
                });
                if (response.status >= 300) {
                    const error = await response.text();
                    console.log(error);
                    return;
                }
                const jsonReturn = await response.json();
                console.log(jsonReturn);
            }} style={{ marginTop: 30 }}>
                Add Class
            </Button>
        </div>
        <hr />
        <div>
            <Button onClick={() => {
                signOut();
            }}>
                Sign out
            </Button>
        </div>
    </>
}