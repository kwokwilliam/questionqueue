import React, { useState, useEffect } from 'react';
import BackToHubButton from '../BackToHubButton';
import { CardDeck } from 'reactstrap';
import PersonInQueue from './Components/PersonInQueue';
import Endpoints from '../../../../../Endpoints/Endpoints';
import cookies from 'browser-cookies';


export default function TutorQAdminWhoIsInQueue() {
    const [queue, setQueue] = useState([]);
    const id = cookies.get('id')

    const uid = localStorage.getItem("Authorization") ? localStorage.getItem("Authorization").split(" ")[1] : "";
    useEffect(() => {
        const { QueueWebSocket } = Endpoints;
        // Connect to websocket here with auth token
        const queueSocket = new WebSocket(`${QueueWebSocket}?identification=${id}&auth=${uid}`)

        queueSocket.onopen = () => {
            queueSocket.send("asdf");
        }

        queueSocket.onmessage = (event) => {
            const { data } = event;
            const parsedData = JSON.parse(data);
            if (parsedData) {
                setQueue(parsedData.queue);
            } else {
                setQueue([])
            }
        }

        return () => {
            queueSocket.close();
        }
    }, []);

    let queueAsArr = queue ? queue.map((d, i) => {
        return <PersonInQueue key={"person" + i} person={d} />
    }) : [];


    return <>
        <BackToHubButton />
        {queueAsArr.length === 0 && <div>There is nobody in the queue right now.</div>}
        {queueAsArr.length !== 0 && <>
            <h3>There are {queueAsArr.length} people in the queue.</h3>
            <CardDeck style={{ textAlign: 'left' }}>{queueAsArr}</CardDeck>
        </>}
    </>
}