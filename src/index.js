import Vue from 'vue'
import axios from 'axios'
import * as Cookies from 'js-cookie';

import { Base } from './page/base';
import { YlaDialog } from './component/yla-dialog';
import { YlaTooltip } from './component/yla-tooltip';

import './less/stylesheet.less'
import 'typeface-source-sans-pro'
import '@fortawesome/fontawesome-free/css/all.css'

// Define global variable
var pageSize = 30;

// Prepare axios instance
var token = Cookies.get('token'),
    rest = axios.create();

rest.defaults.timeout = 60000;
rest.defaults.headers.common['Authorization'] = 'Bearer ' + token;

// Register Vue component
Vue.component('yla-dialog', new YlaDialog());
Vue.component('yla-tooltip', new YlaTooltip());

new Vue({
    el: '#index-page',
    mixins: [new Base()],
    data() {
        return {
            loading: false,
            tags: [],
            bookmarks: [],
            search: '',
            page: 0,
            maxPage: 0,
            editMode: false,
            selected: [],
            bookmarklet: '',
            options: {
                listView: false,
                nightMode: false,
                showBookmarkID: false,
                mainOpenOriginal: false,
            },
            dialogAbout: {
                visible: false,
                title: 'About',
                mainClick: () => {
                    this.dialogAbout.visible = false;
                },
            },
            dialogTags: {
                visible: false,
                loading: false,
                title: 'Existing Tags',
                mainText: 'Cancel',
                mainClick: () => {
                    this.dialogTags.visible = false;
                },
            },
            dialogOptions: {
                visible: false,
                title: 'Options',
                mainText: 'OK',
                mainClick: () => {
                    this.dialogOptions.visible = false;
                },
            }
        }
    },
    computed: {
        visibleBookmarks() {
            var start = this.page * pageSize,
                finish = start + pageSize;
            if (this.bookmarks) {
                return this.bookmarks.slice(start, finish);
            }
            return []
        }
    },
    methods: {
        loadData() {
            if (this.loading) return;

            // Parse search query
            var rxTagA = /['"]#([^'"]+)['"]/g,
                rxTagB = /(^|\s+)#(\S+)/g,
                keyword = this.search,
                tags = [],
                result = [];

            // Fetch tag A first
            while ((result = rxTagA.exec(keyword)) !== null) {
                tags.push(result[1]);
            }

            // Clear tag A from keyword
            keyword = keyword.replace(rxTagA, '');

            // Fetch tag B
            while ((result = rxTagB.exec(keyword)) !== null) {
                tags.push(result[2]);
            }

            // Clear tag B from keyword and clean it
            keyword = keyword.replace(rxTagB, '').trim().replace(/\s+/g, ' ');

            // Fetch data
            this.loading = true;
            rest.get('/api/bookmarks', {
                    params: {
                        keyword: keyword,
                        tags: tags.join(',')
                    }
                })
                .then((response) => {
                    this.page = 0;
                    this.bookmarks = response.data;
                    this.maxPage = Math.ceil(this.bookmarks.length / pageSize) - 1;
                    window.scrollTo(0, 0);

                    return rest.get('/api/tags');
                })
                .then((response) => {
                    this.tags = response.data;
                    this.loading = false;
                })
                .catch((error) => {
                    this.loading = false;

                    var errorMsg = (error.response ? error.response.data : error.message).trim();
                    if (errorMsg.startsWith('Token error')) this.showDialogSessionExpired(errorMsg);
                    else this.showErrorDialog(errorMsg);
                });
        },
        reloadData() {
            if (this.loading) return;
            this.search = '';
            this.loadData();
        },
        changePage(target) {
            target = parseInt(target, 10) || 0;

            if (target >= this.maxPage) this.page = this.maxPage;
            else if (target <= 0) this.page = 0;
            else this.page = target;

            window.scrollTo(0, 0);
        },
        toggleListView() {
            this.options.listView = !this.options.listView;
            window.scrollTo(0, 0);
            localStorage.setItem('shiori-list-view', this.options.listView ? '1' : '0');
        },
        toggleNightMode() {
            this.options.nightMode = !this.options.nightMode;
            localStorage.setItem('shiori-night-mode', this.options.nightMode ? '1' : '0');
        },
        toggleBookmarkID() {
            this.options.showBookmarkID = !this.options.showBookmarkID;
            localStorage.setItem('shiori-show-id', this.options.showBookmarkID ? '1' : '0');
        },
        toggleBookmarkMainLink() {
            this.options.mainOpenOriginal = !this.options.mainOpenOriginal;
            localStorage.setItem('shiori-main-original', this.options.mainOpenOriginal ? '1' : '0');
        },
        toggleEditMode() {
            this.editMode = !this.editMode;
            this.selected = [];
        },
        toggleSelection(idx) {
            var pos = this.selected.indexOf(idx);
            if (pos === -1) this.selected.push(idx);
            else this.selected.splice(pos, 1);
        },
        isSelected(idx) {
            return this.selected.indexOf(idx) > -1;
        },
        filterTag(tag) {
            // Prepare variable
            var rxSpace = /\s+/g,
                searchTag = rxSpace.test(tag) ? '"#' + tag + '"' : '#' + tag;

            // Check if tag already exist in search
            var rxTag = new RegExp(searchTag, 'g');
            if (rxTag.test(this.search)) return;

            // Create new search query
            var newSearch = this.search
                .replace(rxTag, '')
                .replace(rxSpace, ' ')
                .trim();

            // Load data
            this.search = (newSearch + ' ' + searchTag).trim();
            this.dialogTags.visible = false;
            this.loadData();
        },
        showDialogAdd() {
            this.showDialog({
                title: 'New Bookmark',
                content: 'Create a new bookmark',
                fields: [{
                    name: 'url',
                    label: 'Url, start with http://...',
                }, {
                    name: 'title',
                    label: 'Custom title (optional)'
                }, {
                    name: 'excerpt',
                    label: 'Custom excerpt (optional)',
                    type: 'area'
                }, {
                    name: 'tags',
                    label: 'Comma separated tags (optional)',
                    separator: ',',
                    dictionary: this.tags.map(tag => tag.name)
                }, ],
                mainText: 'OK',
                secondText: 'Cancel',
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

                    // Send data
                    this.dialog.loading = true;
                    rest.post('/api/bookmarks', {
                            url: data.url.trim(),
                            title: data.title.trim(),
                            excerpt: data.excerpt.trim(),
                            tags: tags
                        })
                        .then((response) => {
                            this.dialog.loading = false;
                            this.dialog.visible = false;
                            this.bookmarks.splice(0, 0, response.data);
                        })
                        .catch((error) => {
                            var errorMsg = (error.response ? error.response.data : error.message).trim();
                            if (errorMsg.startsWith('Token error')) this.showDialogSessionExpired(errorMsg);
                            else this.showErrorDialog(errorMsg);
                        });
                }
            });
        },
        showDialogEdit(idx) {
            idx += this.page * pageSize;

            var book = this.bookmarks ? JSON.parse(JSON.stringify(this.bookmarks[idx])) : [],
                strTags = book.tags ? book.tags.map(tag => tag.name).join(', '): '';

            this.showDialog({
                title: 'Edit Bookmark',
                content: 'Edit the bookmark\'s data',
                showLabel: true,
                fields: [{
                    name: 'title',
                    label: 'Title',
                    value: book.title,
                }, {
                    name: 'excerpt',
                    label: 'Excerpt',
                    type: 'area',
                    value: book.excerpt,
                }, {
                    name: 'tags',
                    label: 'Tags',
                    value: strTags,
                }],
                mainText: 'OK',
                secondText: 'Cancel',
                mainClick: (data) => {
                    // Validate input
                    if (data.title.trim() === '') return;

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

                    // Set new data
                    book.title = data.title.trim();
                    book.excerpt = data.excerpt.trim();
                    book.tags = tags;

                    // Send data
                    this.dialog.loading = true;
                    rest.put('/api/bookmarks', book)
                        .then((response) => {
                            this.dialog.loading = false;
                            this.dialog.visible = false;
                            this.bookmarks.splice(idx, 1, response.data);
                        })
                        .catch((error) => {
                            var errorMsg = (error.response ? error.response.data : error.message).trim();
                            if (errorMsg.startsWith('Token error')) this.showDialogSessionExpired(errorMsg);
                            else this.showErrorDialog(errorMsg);
                        });
                }
            });
        },
        showDialogDelete(indices) {
            // Check and prepare indices
            if (!(indices instanceof Array)) return;
            if (indices.length === 0) return;
            indices.sort();

            // Set real indices value
            indices = indices.map(item => item + this.page * pageSize)

            // Create title and content
            var title = "Delete Bookmarks",
                content = "Delete the selected bookmarks ? This action is irreversible.";

            if (indices.length === 1) {
                title = "Delete Bookmark";
                content = "Are you sure ? This action is irreversible.";
            }

            // Get list of bookmark ID
            var listID = [];
            for (var i = 0; i < indices.length; i++) {
                listID.push(this.bookmarks[indices[i]].id);
            }

            // Show dialog
            this.showDialog({
                title: title,
                content: content,
                mainText: 'Yes',
                secondText: 'No',
                mainClick: () => {
                    this.dialog.loading = true;
                    rest.delete('/api/bookmarks/', {
                            data: listID
                        })
                        .then((response) => {
                            this.selected = [];
                            this.editMode = false;
                            this.dialog.loading = false;
                            this.dialog.visible = false;
                            for (var i = indices.length - 1; i >= 0; i--) {
                                this.bookmarks.splice(indices[i], 1);
                            }
                        })
                        .catch((error) => {
                            var errorMsg = (error.response ? error.response.data : error.message).trim();
                            if (errorMsg.startsWith('Token error')) this.showDialogSessionExpired(errorMsg);
                            else this.showErrorDialog(errorMsg);
                        });
                }
            });
        },
        showDialogUpdateCache(indices) {
            // Check and prepare indices
            if (!(indices instanceof Array)) return;
            if (indices.length === 0) return;
            indices.sort();

            // Set real indices value
            indices = indices.map(item => item + this.page * pageSize)

            // Get list of bookmark ID
            var listID = [];
            for (var i = 0; i < indices.length; i++) {
                listID.push(this.bookmarks[indices[i]].id);
            }

            // Show dialog
            this.showDialog({
                title: 'Update Cache',
                content: 'Update cache for selected bookmarks ? This action is irreversible.',
                mainText: 'Yes',
                secondText: 'No',
                mainClick: () => {
                    this.dialog.loading = true;
                    rest.put('/api/cache/', listID)
                        .then((response) => {
                            this.selected = [];
                            this.editMode = false;
                            this.dialog.loading = false;
                            this.dialog.visible = false;

                            response.data.forEach(book => {
                                for (var i = 0; i < indices.length; i++) {
                                    var idx = indices[i];
                                    if (book.id === this.bookmarks[idx].id) {
                                        this.bookmarks.splice(idx, 1, book);
                                        break;
                                    }
                                }
                            });
                        })
                        .catch((error) => {
                            var errorMsg = (error.response ? error.response.data : error.message).trim();
                            if (errorMsg.startsWith('Token error')) this.showDialogSessionExpired(errorMsg);
                            else this.showErrorDialog(errorMsg);
                        });
                }
            });
        },
        showDialogAddTags(indices) {
            // Check and prepare indices
            if (!(indices instanceof Array)) return;
            if (indices.length === 0) return;
            indices.sort();

            // Set real indices value
            indices = indices.map(item => item + this.page * pageSize)

            // Get list of bookmark ID
            var listID = [];
            for (var i = 0; i < indices.length; i++) {
                listID.push(this.bookmarks[indices[i]].id);
            }

            this.showDialog({
                title: 'Add New Tags',
                content: 'Add new tags to selected bookmarks',
                fields: [{
                    name: 'tags',
                    label: 'Comma separated tags',
                    value: '',
                }],
                mainText: 'OK',
                secondText: 'Cancel',
                mainClick: (data) => {
                    // Validate input
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

                    if (tags.length === 0) return;

                    // Send data
                    this.dialog.loading = true;
                    rest.put('/api/bookmarks/tags', {
                            ids: listID,
                            tags: tags,
                        })
                        .then((response) => {
                            this.selected = [];
                            this.editMode = false;
                            this.dialog.loading = false;
                            this.dialog.visible = false;

                            response.data.forEach(book => {
                                for (var i = 0; i < indices.length; i++) {
                                    var idx = indices[i];
                                    if (book.id === this.bookmarks[idx].id) {
                                        this.bookmarks.splice(idx, 1, book);
                                        break;
                                    }
                                }
                            });
                        })
                        .catch((error) => {
                            var errorMsg = (error.response ? error.response.data : error.message).trim();
                            if (errorMsg.startsWith('Token error')) this.showDialogSessionExpired(errorMsg);
                            else this.showErrorDialog(errorMsg);
                        });
                }
            });
        },
        showDialogTags() {
            this.dialogTags.visible = true;
            this.dialogTags.loading = true;
            rest.get('/api/tags', {
                    timeout: 5000
                })
                .then((response) => {
                    this.tags = response.data;
                    this.dialogTags.loading = false;
                })
                .catch((error) => {
                    this.dialogTags.loading = false;
                    this.dialogTags.visible = false;

                    var errorMsg = (error.response ? error.response.data : error.message).trim();
                    if (errorMsg.startsWith('Token error')) this.showDialogSessionExpired(errorMsg);
                    else this.showErrorDialog(errorMsg);
                });
        },
        showDialogLogout() {
            this.showDialog({
                title: 'Log Out',
                content: 'Do you want to log out from shiori ?',
                mainText: 'Yes',
                secondText: 'No',
                mainClick: () => {
                    Cookies.remove('token');
                    location.href = '/login';
                }
            });
        },
        showDialogSessionExpired(msg) {
            this.showDialog({
                title: 'Error',
                content: msg + '. Please login again.',
                mainText: 'OK',
                mainClick: () => {
                    Cookies.remove('token');
                    location.href = '/login';
                }
            });
        },
        showDialogAbout() {
            this.dialogAbout.visible = true;
        },
        showDialogOptions() {
            this.dialogOptions.visible = true;
        },
        getHostname(url) {
            var parser = document.createElement('a');
            parser.href = url;
            return parser.hostname.replace(/^www\./g, '');
        },
        getBookLink(book, isMainLink) {
            if ((this.options.mainOpenOriginal && isMainLink) ||
                (!this.options.mainOpenOriginal && !isMainLink)) return book.url;
            if (book.content.length > 0) return '/bookmark/' + book.id;
            return null;
        },
        getBookLinkTitle(book, isMainLink) {
            if ((this.options.mainOpenOriginal && isMainLink) ||
                (!this.options.mainOpenOriginal && !isMainLink)) return 'View original';
            if (book.content.length > 0) return 'View cache';
            return null;
        }
    },
    mounted() {
        // Read config from local storage
        var listView = localStorage.getItem('shiori-list-view'),
            nightMode = localStorage.getItem('shiori-night-mode'),
            showBookmarkID = localStorage.getItem('shiori-show-id'),
            mainOpenOriginal = localStorage.getItem('shiori-main-original');

        this.options.listView = listView === '1';
        this.options.nightMode = nightMode === '1';
        this.options.showBookmarkID = showBookmarkID === '1';
        this.options.mainOpenOriginal = mainOpenOriginal === '1';

        // Create bookmarklet
        var shioriURL = location.href.replace(/\/+$/g, ''),
            baseBookmarklet = `(function () {
                var shioriURL = '$SHIORI_URL',
                    bookmarkURL = location.href,
                    submitURL = shioriURL + '/submit?url=' + encodeURIComponent(bookmarkURL);

                if (bookmarkURL.startsWith('https://') && !shioriURL.startsWith('https://')) {
                    window.open(submitURL, '_blank');
                    return;
                }

                var i = document.createElement('iframe');
                i.src = submitURL;
                i.frameBorder = '0';
                i.allowTransparency = true;
                i.style.position = 'fixed';
                i.style.top = 0;
                i.style.left = 0;
                i.style.width = '100%';
                i.style.height = '100%';
                i.style.zIndex = 99999;
                document.body.appendChild(i);

                window.addEventListener('message', function onMessage(e) {
                    if (e.origin !== shioriURL) return;
                    if (e.data !== 'finished') return;
                    window.removeEventListener('message', onMessage);
                    document.body.removeChild(i);
                });
            }())`;

        this.bookmarklet = 'javascript:' + baseBookmarklet
            .replace('$SHIORI_URL', shioriURL)
            .replace(/\s+/gm, ' ');

        // Load data
        this.loadData();
    }
});