import React, { Component } from 'react';
import { Button } from 'reactstrap';
import { Route } from 'react-router-dom';
import Spinner from 'react-loader-spinner';
import Loadable from 'react-loadable';
import Endpoints from '../../../Endpoints/Endpoints';

const Loading = () => <div><Spinner
    type="Oval"
    color="#005696"
    height="100"
    width="100"
/></div>;

const TutorQAdminMain = Loadable({
    loader: () => import('./Components/Main/TutorQAdminMain'),
    loading: Loading,
});

const TutorQAdminWhoIsInQueue = Loadable({
    loader: () => import('./Components/WhoInQueue/TutorQAdminWhoIsInQueue'),
    loading: Loading,
});

const TutorQAdminAdminQueue = Loadable({
    loader: () => import('./Components/AdminQueue/TutorQAdminAdminQueue'),
    loading: Loading,
});

const TutorQAdminSeatingDistribution = Loadable({
    loader: () => import('./Components/SeatingDistribution/TutorQAdminSeatingDistribution'),
    loading: Loading,
});


export default class TutorQAdmin extends Component {
    constructor(props) {
        super(props);
        this.state = {
            authToken: localStorage.getItem("Authorization") || null,
            user: null,
            loading: true,
            admin: false
        }

        this.adminButtons = [
            {
                linkTo: "/tutorqadmin/adminqueue",
                linkText: "Main Admin Tool"
            },
            {
                linkTo: "/tutorqadmin/whosinqueue",
                linkText: "See Queue List"
            },
            {
                linkTo: "/tutorqadmin/seatingdistribution",
                linkText: "Seating Distribution"
            },
            // {
            //     linkTo: "/tutorqadmin/statistics",
            //     linkText: "Tutor Statistics"
            // }
        ];

    }

    componentDidMount = async () => {
        if (!this.state.authToken) {
            this.setState({ loading: false });
            return;
        }

        const { URL, Teacher } = Endpoints;
        const response = await fetch(URL + Teacher + "/me", {
            headers: new Headers({
                "Authorization": this.state.authToken
            })
        });
        if (response.status >= 300) {
            alert("Unable to verify login. Logging out...");
            localStorage.setItem("Authorization", "");
            this.setAuthToken("");
            this.setUser(null);
            return;
        }
        const user = await response.json()
        this.setUser(user);
    }

    setAuthToken = (authToken) => {
        this.setState({ authToken });
    }

    setUser = (user) => {
        this.setState({ user, admin: user ? user.admin : false, loading: false })
    }

    signOut = async () => {
        const { URL, TeacherLogin } = Endpoints;
        const response = await fetch(URL + TeacherLogin, {
            method: "DELETE",
            headers: new Headers({
                "Authorization": this.state.authToken
            })
        });
        if (response.status >= 300) {
            alert("Unable to sign out");
            return;
        }
        this.setUser(null);
        this.setAuthToken(null);
    }

    render() {
        const { loading, user, admin } = this.state;
        console.log(loading);
        return <div style={{ textAlign: 'center' }}>
            <h1 style={{ marginBottom: '5vh' }}>
                TutorQ Admin Panel
            </h1>

            {loading && <Loading />}

            {!loading && !user && <div>
                <Button onClick={() => {
                    // firebase.auth().signInWithRedirect(provider);
                }} style={{ backgroundColor: "#005696" }}>Sign in</Button>
            </div>}

            {user && admin && <>
                <Route exact path={"/tutorqadmin"} render={() => <TutorQAdminMain adminButtons={this.adminButtons} />} />
                <Route path={"/tutorqadmin/adminqueue"} render={() => <TutorQAdminAdminQueue uid={user.id} />} />
                <Route path={"/tutorqadmin/whosinqueue"} render={() => <TutorQAdminWhoIsInQueue />} />
                <Route path={"/tutorqadmin/seatingdistribution"} render={() => <TutorQAdminSeatingDistribution />} />
            </>}

            {user && !admin && <>
                <h1>You are not permitted to view this page.</h1>
                <Button onClick={() => {
                    this.signOut();
                }}>
                    Sign out
                </Button>
            </>}
        </div>
    }
}

// TODO: CONCURRENCY AND SLOW INTERNET