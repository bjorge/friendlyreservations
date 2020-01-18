import React, { Component } from 'react';

import Header from './Header'
import Main from './Main'
import { instanceOf } from 'prop-types';
import { withCookies, Cookies } from 'react-cookie';


class AuthCookies extends Component {
    static propTypes = {
        cookies: instanceOf(Cookies).isRequired
    };

    constructor(props) {
        super(props);

        this.isAuthenticated = this.isAuthenticated.bind(this);
        this.handleLogout = this.handleLogout.bind(this);
    }

    handleLogout() {
        console.log("handleLogout, clear cookies")
        this.props.cookies.remove("jsauth", {});
    }

    isAuthenticated() {
        console.log("render, loginCheck");
        let sessionCookie = this.props.cookies.get('jsauth');
        if (sessionCookie === undefined) {
            console.log("AuthCookies: session cookie is undefined, go to /login page");
            // localStorage.clear();
            return false;
        }
        // console.log("sessionCookie:");
        // console.log(sessionCookie);

        var json = atob(sessionCookie);
        console.log("AuthCookies: session json from server:")
        console.log(json);

        var value = JSON.parse(json);
        // console.log(value);

        var expiration = new Date(value.expiration);
        // console.log(expiration.toLocaleString());

        var now = new Date();
        // console.log(now.toLocaleString());

        if (expiration < now) {
            console.log("AuthCookies: session cookie has expired, go to /login page");
            localStorage.clear();
            return false;
        }

        // localStorage.setItem(AUTH_TOKEN, sessionCookie);
        // localStorage.setItem(CURR_NUMSCHI, value.numschi);
        // localStorage.setItem(IS_ADMIN, value.isAdmin);

        console.log("AuthCookies: session is valid")
        return true;
    }

    render() {
        const authenticated = this.isAuthenticated();
        return (
            <div>
                <Header authenticated={authenticated} logoutCallback={this.handleLogout} />
                <Main authenticated={authenticated} />
            </div>);
    }

}
export default withCookies(AuthCookies);

