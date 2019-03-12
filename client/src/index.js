import React from 'react';
import ReactDOM from 'react-dom';
import App from './App';
import * as serviceWorker from './serviceWorker';
import crypto from 'crypto';
import cookies from 'browser-cookies';

// Print version number
console.log("v1.0.0");

// Set up unique id 
let id = cookies.get('id');
if (!id) {
    id = crypto.randomBytes(64).toString('hex');
    cookies.set('id', id);
}


ReactDOM.render(<App />, document.getElementById('root'));

serviceWorker.unregister();
