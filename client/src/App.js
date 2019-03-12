import React, { Component } from 'react';
import {
    HashRouter as Router, // Would like to ideally use nextjs but hash router is good enough for this project.
    Switch,
    Route,
    Redirect
} from 'react-router-dom';
import cookies from 'browser-cookies';
import TutorQStudent from './Pages/TutorQ/Student/TutorQStudent';
import './App.css';


class App extends Component {
    state = {
        id: cookies.get('id')
    }

    render() {
        const { id } = this.state;
        return (
            <>
                <Router>
                    <Switch>
                        <Route exact path="/" component={() => <TutorQStudent />} />
                    </Switch>
                </Router>
            </>
        );
    }
}

export default App;
