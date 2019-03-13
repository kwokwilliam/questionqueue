import React, { Component } from 'react';
import { Input, Button } from 'reactstrap';
import { Route, Switch } from 'react-router-dom';
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

// const TutorQAdminAdminQueue = Loadable({
//     loader: () => import('./Components/AdminQueue/TutorQAdminAdminQueue'),
//     loading: Loading,
// });

// const TutorQAdminSeatingDistribution = Loadable({
//     loader: () => import('./Components/SeatingDistribution/TutorQAdminSeatingDistribution'),
//     loading: Loading,
// });


export default class TutorQAdmin extends Component {
    constructor(props) {
        super(props);
        this.state = {
            authToken: localStorage.getItem("Authorization") || null,
            user: null,
            loading: true,
            admin: false,
            loginEmail: "",
            loginPassword: "",
            signUpEmail: "",
            signUpPassword: "",
            signUpPasswordConf: "",
            signUpFirstName: "",
            signUpLastName: ""
        }

        this.adminButtons = [
            // {
            //     linkTo: "/tutorqadmin/adminqueue",
            //     linkText: "Main Admin Tool"
            // },
            {
                linkTo: "/tutorqadmin/whosinqueue",
                linkText: "See Queue List"
            },
            // {
            //     linkTo: "/tutorqadmin/seatingdistribution",
            //     linkText: "Seating Distribution"
            // },
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

    signIn = async () => {
        const { URL, TeacherLogin } = Endpoints;
        const { loginEmail, loginPassword } = this.state;
        const sendData = {
            email: loginEmail,
            password: loginPassword
        }
        const response = await fetch(URL + TeacherLogin, {
            method: "POST",
            body: JSON.stringify(sendData),
            headers: new Headers({
                "Content-Type": "application/json"
            })
        });
        if (response.status >= 300) {
            const error = await response.text();
            this.setError(error);
            return;
        }
        const authToken = response.headers.get("Authorization")
        localStorage.setItem("Authorization", authToken);
        this.setError("");
        this.setAuthToken(authToken);
        const user = await response.json();
        this.setUser(user);
        this.setState({
            loginEmail: "",
            loginPassword: ""
        });
    }

    signUp = async () => {
        const { URL, Teacher } = Endpoints;
        const { signUpEmail, signUpPassword, signUpPasswordConf, signUpFirstName, signUpLastName } = this.state;
        const sendData = {
            email: signUpEmail,
            password: signUpPassword,
            password_conf: signUpPasswordConf,
            firstname: signUpFirstName,
            lastname: signUpLastName
        }
        const response = await fetch(URL + Teacher, {
            method: "POST",
            body: JSON.stringify(sendData),
            headers: new Headers({
                "Content-Type": "application/json"
            })
        });
        if (response.status >= 300) {
            const error = await response.text();
            this.setError(error);
            return;
        }
        const authToken = response.headers.get("Authorization");
        localStorage.setItem("Authorization", authToken);
        this.setError("");
        this.setAuthToken(authToken);
        const user = await response.json();
        this.setUser(user);
        this.setState({
            signUpEmail: "",
            signUpPassword: "",
            signUpPasswordConf: "",
            signUpFirstName: "",
            signUpLastName: ""
        })
    }

    change = (e) => {
        this.setState({
            [e.target.name]: e.target.value
        });
    }

    render() {
        const { loading, user, admin } = this.state;
        const inputStyles = { maxWidth: 500, margin: 'auto', marginTop: 30 };
        return <div style={{ textAlign: 'center' }}>
            <h1 style={{ marginBottom: '5vh' }}>
                QuestionQueue Admin Panel
            </h1>

            {loading && <Loading />}

            {!loading && !user && <div>
                <h2>Sign In</h2>
                <Input placeholder={"Email"}
                    name="loginEmail"
                    onChange={this.change}
                    value={this.state.loginEmail}
                    style={inputStyles}
                />
                <Input placeholder={"Password"}
                    name="loginPassword"
                    onChange={this.change}
                    type="password"
                    value={this.state.loginPassword}
                    style={inputStyles}
                />
                <Button onClick={() => {
                    this.signIn();
                }} style={{ backgroundColor: "#005696", marginTop: 10 }}>Sign in</Button>
                <div style={{ marginBottom: 20 }}></div>
                <h2>Sign Up</h2>
                <Input placeholder={"Email"}
                    name="signUpEmail"
                    onChange={this.change}
                    value={this.state.signUpEmail}
                    style={inputStyles}
                />
                <Input placeholder={"Password"}
                    name="signUpPassword"
                    type="password"
                    onChange={this.change}
                    value={this.state.signUpPassword}
                    style={inputStyles}
                />
                <Input placeholder={"Password confirmation"}
                    name="signUpPasswordConf"
                    type="password"
                    onChange={this.change}
                    value={this.state.signUpPasswordConf}
                    style={inputStyles}
                />
                <Input placeholder={"First Name"}
                    name="signUpFirstName"
                    onChange={this.change}
                    value={this.state.signUpFirstName}
                    style={inputStyles}
                />
                <Input placeholder={"Last Name"}
                    name="signUpLastName"
                    onChange={this.change}
                    value={this.state.signUpLastName}
                    style={inputStyles}
                />
                <Button onClick={() => {
                    this.signUp();
                }} style={{ backgroundColor: "#005696", marginTop: 10 }}>Sign up</Button>

            </div>}

            {user && admin && <>
                <Route exact path={"/tutorqadmin"} render={() => <TutorQAdminMain adminButtons={this.adminButtons} signOut={this.signOut} />} />
                {/* <Route path={"/tutorqadmin/adminqueue"} render={() => <TutorQAdminAdminQueue uid={user.id} identification={}/>} /> */}
                <Route path={"/tutorqadmin/whosinqueue"} render={() => <TutorQAdminWhoIsInQueue uid={user.id} />} />
                {/* <Route path={"/tutorqadmin/seatingdistribution"} render={() => <TutorQAdminSeatingDistribution />} /> */}
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