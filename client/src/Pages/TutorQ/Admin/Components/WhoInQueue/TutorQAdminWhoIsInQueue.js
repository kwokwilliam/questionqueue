import React, { useState, useEffect } from 'react';
import BackToHubButton from '../BackToHubButton';
import { CardDeck } from 'reactstrap';
import PersonInQueue from './Components/PersonInQueue';
import Endpoints from '../../../../../Endpoints/Endpoints';


export default function TutorQAdminWhoIsInQueue() {
    const [queue, setQueue] = useState([]);

    useEffect(() => {
        const { QueueWebSocket } = Endpoints;
        // Connect to websocket here with auth token
        const queueSocket = new WebSocket(`${QueueWebSocket}?identification=${this.id}&auth=${this.state.uid}`)

        queueSocket.onopen = () => {
            console.log("Connected");
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

    let queueAsArr = queue.map((d, i) => {
        return <PersonInQueue key={"person" + i} person={d} />
    })

    return <>
        <BackToHubButton />
        {queueAsArr.length === 0 && <div>There is nobody in the queue right now.</div>}
        {queueAsArr.length !== 0 && <>
            <h3>There are {queueAsArr.length} people in the queue.</h3>
            <CardDeck style={{ textAlign: 'left' }}>{queueAsArr}</CardDeck>
        </>}
    </>
}