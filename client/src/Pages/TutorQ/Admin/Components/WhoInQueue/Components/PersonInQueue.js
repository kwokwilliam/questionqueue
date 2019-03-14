import React, { useState } from 'react';
import {
    Card, CardText, CardBody,
    CardTitle, CardSubtitle, Button
} from 'reactstrap';
import Endpoints from '../../../../../../Endpoints/Endpoints';
import StudentLocation from '../../../../Components/StudentLocation/StudentLocation';


// TODO: Use an effect to stop setLoading from happening when component unmounts
export default function PersonInQueue({ person }) {
    const [loading, setLoading] = useState(false);

    const id = person.id;
    const name = person.name;
    const classNumber = person['class'];
    const problemCategory = person.topic;
    const problemDescription = person.problem;
    const location = {
        xPercentage: person["loc_x"],
        yPercentage: person["loc_y"]
    }
    const timestamp = person.created_at;
    console.log("abc2");

    if (!timestamp) { return null; }
    let dateTimestamp = new Date(timestamp);
    return <Card>
        <CardBody>
            <CardTitle>Name: {name}</CardTitle>
            <CardSubtitle>Course: {classNumber} - {problemCategory}</CardSubtitle>
            <CardText>Submitted: {dateTimestamp.toLocaleDateString()} {dateTimestamp.toLocaleTimeString()}</CardText>
            <CardText>Description: {problemDescription}</CardText>
            <StudentLocation locations={[location]} student={false} />
            <Button disabled={loading} style={{ backgroundColor: "#005696" }} onClick={async () => {
                setLoading(true);
                const { URL, Queue } = Endpoints;
                const response = await fetch(URL + Queue + "/" + id, {
                    method: "DELETE"
                });
                if (response.status >= 300) {
                    const error = await response.text();
                    console.log(error);
                    return;
                }
                setLoading(false);
                console.log("queue removal successful");
            }}>Remove</Button>
        </CardBody>
    </Card>
}