import Vue from 'vue'

import './sass/stylesheet.scss'

new Vue({
    el: '#cache-page',
    data: {
        nightMode: false,
        serifMode: false,
        html: window.__INIT__.html,
        url: window.__INIT__.url,
        title: window.__INIT__.title,
    },
    methods: {
        toggleNightMode() {
            this.nightMode = !this.nightMode;
            localStorage.setItem('shiori-night-mode', this.nightMode ? '1' : '0');
        },
    },
    mounted() {
        // Set title
        document.title = this.title + ' - Shiori - Bookmarks Manager';

        // Set night and serif mode
        var nightMode = localStorage.getItem('shiori-night-mode'),
            serifMode = localStorage.getItem('shiori-serif-mode');

        this.nightMode = nightMode === '1';
    }
});