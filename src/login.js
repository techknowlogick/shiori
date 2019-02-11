import Vue from 'vue'
import axios from 'axios'
import * as Cookies from 'js-cookie';

import 'typeface-source-sans-pro'
import '@fortawesome/fontawesome-free/css/all.css'

new Vue({
  el: '#login-page',
  data: {
    nightMode: true,
    error: '',
    loading: false,
    username: '',
    password: '',
    rememberMe: false,
  },
  methods: {
        toggleRemember: function () {
            this.rememberMe = !this.rememberMe;
        },
        login: function () {
            // Validate input
            if (this.username === '') {
                this.error = 'Username must not empty';
                return;
            }

            // Send request
            this.loading = true;
            axios.post('/api/login', {
                    username: this.username,
                    password: this.password,
                    remember: this.rememberMe
                }, {
                    timeout: 10000
                })
                .then(function (response) {
                    // Save token
                    var token = response.data;
                    Cookies.set('token', token);

                    // Set destination URL
                    var rx = /[&?]dst=([^&]+)(&|$)/g,
                        match = rx.exec(location.href);

                    if (match == null) {
                        location.href = '/';
                    } else {
                        var dst = match[1];
                        location.href = decodeURIComponent(dst);
                    }
                })
                .catch(function (error) {
                    var errorMsg = (error.response ? error.response.data : error.message).trim();
                    app.password = '';
                    app.loading = false;
                    app.error = errorMsg;
                });
        }
    },
    mounted() {
        // Read config from local storage
        var nightMode = localStorage.getItem('shiori-night-mode');
        this.nightMode = nightMode === '1';
    }
})
