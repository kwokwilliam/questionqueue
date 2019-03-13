import React from 'react';
import { Button } from 'reactstrap';
import { Link } from 'react-router-dom';

export default function TutorQAdminMain({ adminButtons, signOut }) {
    return <>
        <div style={{ fontSize: '150%' }}>
            {adminButtons.map(d => {
                return <div key={d.linkTo} style={{ marginBottom: 5 }}>
                    <Link className="btn" style={{ textDecoration: 'none', color: 'white', backgroundColor: '#005696' }} to={d.linkTo}>{d.linkText}</Link>
                </div>
            })}
        </div>
        <div>
            <Button onClick={() => {
                signOut();
            }}>
                Sign out
            </Button>
        </div>
    </>
}