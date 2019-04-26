import Vue from 'vue'
import axios from 'axios'
import * as Cookies from 'js-cookie';

import { Base } from './page/base';
import { YlaDialog } from './component/yla-dialog';

import './less/stylesheet.less'
import 'typeface-source-sans-pro'
import '@fortawesome/fontawesome-free/css/all.css'

// Create private function
function _inIframe() {
    try {
        return window.self !== window.top;
    } catch (e) {
        return true;
    }
}

// Register Vue component
Vue.component('yla-dialog', new YlaDialog());

// Prepare axios instance
var token = Cookies.get('token'),
    rest = axios.create();

rest.defaults.timeout = 60000;
rest.defaults.headers.common['Authorization'] = 'Bearer ' + token;

new Vue({
    el: '#submit-page',
    mixins: [new Base()],
    data: {
        targetURL: '',
        nightMode: false,
        inIframe: false,
    },
    methods: {
        showDialogLogin() {
            this.showDialog({
                title: 'Login',
                content: 'Please input username and password',
                fields: [{
                    name: 'username',
                    label: 'Username',
                }, {
                    name: 'password',
                    label: 'Password',
                    type: 'password'
                }],
                mainText: 'OK',
                secondText: this.inIframe ? 'Cancel' : '',
                mainClick: (data) => {
                    // Validate input
                    if (data.username.trim() === '') return;

                    // Send data
                    this.dialog.loading = true;
                    rest.post('/api/login', {
                            username: data.username.trim(),
                            password: data.password,
                        })
                        .then((response) => {
                            // Save token
                            var token = response.data;
                            Cookies.set('token', token);

                            this.showDialogAdd();
                        })
                        .catch((error) => {
                            var errorMsg = (error.response ? error.response.data : error.message).trim();
                            this.showErrorDialog(errorMsg);
                            this.dialog.mainClick = () => {
                                this.showDialogLogin();
                            }
                        });
                },
                secondClick: () => {
                    window.top.postMessage('finished', '*');
                    this.dialog.visible = false;
                }
            });
        },
        showDialogAdd() {
            this.showDialog({
                title: 'New Bookmark',
                content: 'Create a new bookmark',
                fields: [{
                    name: 'url',
                    label: 'Url, start with http://...',
                    value: this.targetURL,
                }, {
                    name: 'title',
                    label: 'Custom title (optional)'
                }, {
                    name: 'excerpt',
                    label: 'Custom excerpt (optional)',
                    type: 'area'
                }, {
                    name: 'tags',
                    label: 'Comma separated tags (optional)'
                }, ],
                mainText: 'Save',
                secondText: this.inIframe ? 'Cancel' : '',
                mainClick: (data) => {
                    // Prepare tags
                    var tags = data.tags
                        .toLowerCase()
                        .replace(/\s+/g, ' ')
                        .split(/\s*,\s*/g)
                        .filter(tag => tag !== '')
                        .map(tag => {
                            return {
                                name: tag
                            };
                        });

                    // Validate input
                    if (data.url.trim() === '') return;

                    // Send data
                    this.dialog.loading = true;
                    rest.post('/api/bookmarks', {
                            url: data.url.trim(),
                            title: data.title.trim(),
                            excerpt: data.excerpt.trim(),
                            tags: tags
                        })
                        .then((response) => {
                            if (!this.inIframe) {
                                location.href = '/';
                                return;
                            }

                            this.showDialog({
                                title: 'New Bookmark',
                                content: 'The new bookmark has been saved successfully.',
                                mainText: 'OK',
                                mainClick: () => {
                                    window.top.postMessage('finished', '*');
                                    this.dialog.visible = false;
                                }
                            });
                        })
                        .catch((error) => {
                            var errorMsg = (error.response ? error.response.data : error.message);
                            if (errorMsg.startsWith('Token error:')) {
                                this.showDialogLogin();
                                return;
                            }

                            this.showErrorDialog(errorMsg);
                            this.dialog.mainClick = () => {
                                this.showDialogAdd();
                            }
                        });
                },
                secondClick: () => {
                    window.top.postMessage('finished', '*');
                    this.dialog.visible = false;
                }
            });
        }
    },
    mounted() {
        // Check if in iframe
        this.inIframe = _inIframe();

        // Read config from local storage
        var nightMode = localStorage.getItem('shiori-night-mode');
        this.nightMode = nightMode === '1';

        // Get target URL
        var rxURL = /[&?]url=([^&]+)(&|$)/g,
            match = rxURL.exec(location.href);

        if (match != null) {
            var dst = match[1];
            this.targetURL = decodeURIComponent(dst);
        } else {
            this.targetURL = '';
        }

        // Show dialog
        this.showDialogAdd();
    }
});