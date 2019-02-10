import Vue from 'vue'
import axios from 'axios'
import Cookies from 'js-cookie'


 // Create initial variable
var init = JSON.parse('$|.|$');

// Register Vue component
Vue.component('yla-dialog', new YlaDialog());
Vue.component('yla-tooltip', new YlaTooltip());

new Vue({
    el: '#cache-page',
    mixins: [new Base()],
    data: {
        id: init.id,
        url: init.url,
        title: init.title,
        author: init.author,
        minReadTime: init.minReadTime,
        maxReadTime: init.maxReadTime,
        modified: init.modified,
        html: init.html,
        tags: init.tags,
        nightMode: false,
        serifMode: false,
    },
    methods: {
        toggleNightMode() {
            this.nightMode = !this.nightMode;
            localStorage.setItem('shiori-night-mode', this.nightMode ? '1' : '0');
        },
        toggleSerifMode() {
            this.serifMode = !this.serifMode;
            localStorage.setItem('shiori-serif-mode', this.serifMode ? '1' : '0');
        },
        getHostname(url) {
            parser = document.createElement('a');
            parser.href = url;
            return parser.hostname.replace(/^www\./g, '');
        }
    },
    mounted() {
        // Set title
        document.title = this.title + ' - Shiori - Bookmarks Manager';

        // Set night and serif mode
        var nightMode = localStorage.getItem('shiori-night-mode'),
            serifMode = localStorage.getItem('shiori-serif-mode');

        this.nightMode = nightMode === '1';
        this.serifMode = serifMode === '1';
    }
});