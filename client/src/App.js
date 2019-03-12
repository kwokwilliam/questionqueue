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
import TutorQAdmin from './Pages/TutorQ/Admin/TutorQAdmin';


class App extends Component {
    state = {
        id: cookies.get('id')
    }

    render() {
        const { id } = this.state;
        return (
            <>
                <div style={{ margin: 10 }}>
                    <Router>
                        <Switch>
                            <Route exact path="/" component={() => <TutorQStudent />} />
                            <Route exact path="/tutorqadmin" component={() => <TutorQAdmin />} />
                            <Route render={() => <div style={{ margin: 20 }}>Error: Page not found :(</div>} />
                        </Switch>
                    </Router>
                </div>
            </>
        );
    }
}

export default App;
