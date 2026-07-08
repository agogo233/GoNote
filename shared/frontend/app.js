// GoNote Frontend Application

// Configuration constants
const CONFIG = {
    AUTOSAVE_DELAY: 1000,              // ms - Delay before triggering autosave
    SEARCH_DEBOUNCE_DELAY: 500,        // ms - Delay before running note search while typing
    SAVE_INDICATOR_DURATION: 2000,     // ms - How long to show "saved" indicator
    SCROLL_SYNC_DELAY: 50,             // ms - Delay to prevent scroll sync interference
    SCROLL_SYNC_MAX_RETRIES: 10,       // Maximum attempts to find editor/preview elements
    SCROLL_SYNC_RETRY_INTERVAL: 100,   // ms - Time between setupScrollSync retries
    MAX_UNDO_HISTORY: 50,              // Maximum number of undo steps to keep
    DEFAULT_SIDEBAR_WIDTH: 256,        // px - Default sidebar width (w-64 in Tailwind)
    HOMEPAGE_MAX_NOTES: 50,            // Maximum notes to show on homepage grid
    MAX_SAVE_RETRIES: 3,               // Maximum retry attempts for save operations
    SAVE_RETRY_BASE_DELAY: 1000,       // ms - Base delay for retry (exponential backoff)
    NOTES_PER_FOLDER_PAGE: 50,         // Max notes rendered per folder page in sidebar
};

// Expose HOMEPAGE_MAX_NOTES globally for Alpine.js template access
const HOMEPAGE_MAX_NOTES = CONFIG.HOMEPAGE_MAX_NOTES;

// ============================================
// NON-REACTIVE CACHE - 非响应式缓存（组件外部）
// 完全脱离 Alpine.js 响应式追踪，减少内存占用和提升性能
// ============================================
const NoteCache = {
    // Note lookup maps for O(1) wikilink resolution
    noteLookup: {
        byPath: new Map(),           // path -> true
        byPathLower: new Map(),      // path.toLowerCase() -> true
        byName: new Map(),           // name (without .md) -> true  
        byNameLower: new Map(),      // name.toLowerCase() -> true
        byEndPath: new Map(),        // '/filename' and '/filename.md' -> true
    },
    
    // Media lookup map for O(1) media wikilink resolution
    mediaLookup: new Map(),
    
    // Shared note paths cache
    sharedNotePaths: new Set(),
    
    // Homepage cache
    homepageCache: {
        folderPath: null,
        notes: null,
        folders: null,
        breadcrumb: null
    },
    
    // DOM element cache
    domCache: {
        editor: null,
        previewContainer: null,
        previewContent: null,
        sidebar: null,
        themeColorMeta: null
    },
    
    // Preview rendering cache
    previewDebounceTimeout: null,
    lastRenderedContent: '',
    cachedRenderedHTML: '',
    mathDebounceTimeout: null,
    mermaidDebounceTimeout: null,
    
    // Note metadata cache
    lastFrontmatter: null,
    
    // Video initialization tracking
    initializedVideoSources: new Set(),
    
    // Scroll position memory for each note
    scrollPositions: new Map(),
    
    // Get scroll position for a note
    getScrollPosition(notePath) {
        return this.scrollPositions.get(notePath) || { editor: 0, preview: 0 };
    },
    
    // Set scroll position for a note
    setScrollPosition(notePath, editorScroll, previewScroll) {
        this.scrollPositions.set(notePath, { editor: editorScroll, preview: previewScroll });
    },
    
    // Clear all note-related caches (call when notes are reloaded)
    clearNoteCaches() {
        this.noteLookup.byPath.clear();
        this.noteLookup.byPathLower.clear();
        this.noteLookup.byName.clear();
        this.noteLookup.byNameLower.clear();
        this.noteLookup.byEndPath.clear();
        this.mediaLookup.clear();
    },
    
    // Reset homepage cache
    resetHomepageCache() {
        this.homepageCache = {
            folderPath: null,
            notes: null,
            folders: null,
            breadcrumb: null
        };
    },
    
    // Reset shared paths
    resetSharedPaths() {
        this.sharedNotePaths = new Set();
    },
    
    // Reset DOM cache
    resetDomCache() {
        this.domCache = {
            editor: null,
            previewContainer: null,
            previewContent: null,
            sidebar: null,
            themeColorMeta: null
        };
    }
};

// CSRF token utilities for secure POST/DELETE/PUT requests
const CSRF = {
    /**
     * Get CSRF token from cookie (set by server's csrf middleware)
     * @returns {string} CSRF token or empty string if not found
     */
    getToken() {
        const cookies = document.cookie.split(';');
        for (const cookie of cookies) {
            const [name, value] = cookie.trim().split('=');
            if (name === 'csrf_') {
                return decodeURIComponent(value);
            }
        }
        return '';
    },
    
    /**
     * Check if a request method requires CSRF protection
     * @param {string} method - HTTP method
     * @returns {boolean}
     */
    requiresProtection(method) {
        return ['POST', 'PUT', 'DELETE', 'PATCH'].includes(method.toUpperCase());
    },
    
    /**
     * Add CSRF header to fetch options if needed
     * @param {Object} options - fetch options object
     * @returns {Object} Modified options with CSRF header
     */
    addHeader(options = {}) {
        const method = (options.method || 'GET').toUpperCase();
        if (!this.requiresProtection(method)) {
            return options;
        }
        
        const token = this.getToken();
        if (!token) {
            console.warn('CSRF token not found in cookie');
            return options;
        }
        
        return {
            ...options,
            headers: {
                ...options.headers,
                'X-CSRF-Token': token
            }
        };
    }
};

/**
 * Secure fetch wrapper that automatically adds CSRF token for mutating requests
 * @param {string} url - Request URL
 * @param {Object} options - fetch options
 * @returns {Promise<Response>}
 */
async function secureFetch(url, options = {}) {
    return fetch(url, CSRF.addHeader(options));
}

// localStorage settings configuration - centralized definition of all persisted settings
// Each setting has a 'target' field that specifies where the value should be stored in the state
const LOCAL_SETTINGS = {
    // Boolean settings
    syntaxHighlightEnabled: { key: 'syntaxHighlightEnabled', type: 'boolean', default: false, target: 'ui.syntaxHighlightEnabled' },
    readableLineLength: { key: 'readableLineLength', type: 'boolean', default: true, target: 'ui.readableLineLength' },
    favoritesExpanded: { key: 'favoritesExpanded', type: 'boolean', default: true, target: '_favoritesState.expanded' },
    tagsExpanded: { key: 'tagsExpanded', type: 'boolean', default: false, target: 'tags.expanded' },
    hideUnderscoreFolders: { key: 'hideUnderscoreFolders', type: 'boolean', default: false, target: 'ui.hideUnderscoreFolders' },
    // Number settings with validation
    sidebarWidth: { key: 'sidebarWidth', type: 'number', default: CONFIG.DEFAULT_SIDEBAR_WIDTH, min: 200, max: 600, target: 'ui.sidebarWidth' },
    editorWidth: { key: 'editorWidth', type: 'number', default: 50, min: 20, max: 80, target: 'ui.editorWidth' },
    // String settings with validation
    viewMode: { key: 'viewMode', type: 'string', default: 'split', valid: ['edit', 'split', 'preview'], target: 'ui.viewMode' },
    // JSON settings
    favorites: { key: 'noteFavorites', type: 'json', default: [], target: '_favoritesState.list' },
    expandedFolders: { key: 'expandedFolders', type: 'json', default: [], target: 'folders.expanded' },
};

// Centralized error handling
const ErrorHandler = {
    /**
     * Handle errors consistently across the app
     * @param {string} operation - The operation that failed (e.g., "load notes", "save note")
     * @param {Error} error - The error object
     * @param {boolean} showAlert - Whether to show an alert to the user
     */
    handle(operation, error, showAlert = true) {
        // Always log to console for debugging
        console.error(`Failed to ${operation}:`, error);
        
        // Show user-friendly alert if requested
        if (showAlert) {
            // Use custom modal if available, fallback to native alert
            if (window.showAppAlert) {
                window.showAppAlert(`Failed to ${operation}. Please try again.`);
            } else {
                alert(`Failed to ${operation}. Please try again.`);
            }
        }
    }
};

/**
 * Centralized filename validation
 * Supports Unicode characters (international text) but blocks dangerous filesystem characters.
 * Does NOT silently modify filenames - validates and returns status.
 */
const FilenameValidator = {
    // Characters that are forbidden in filenames across Windows/macOS/Linux
    // Windows: \ / : * ? " < > |
    // macOS: / :
    // Linux: / \0
    // Common set to block (including control characters)
    FORBIDDEN_CHARS: /[\\/:*?"<>|\x00-\x1f]/,
    
    // For display purposes - human readable list
    FORBIDDEN_CHARS_DISPLAY: '\\ / : * ? " < > |',
    
    /**
     * Validate a filename (single segment, no path separators)
     * @param {string} name - The filename to validate
     * @returns {{ valid: boolean, error?: string, sanitized?: string }}
     */
    validateFilename(name) {
        if (!name || typeof name !== 'string') {
            return { valid: false, error: 'empty' };
        }
        
        const trimmed = name.trim();
        if (!trimmed) {
            return { valid: false, error: 'empty' };
        }
        
        // Check for forbidden characters
        if (this.FORBIDDEN_CHARS.test(trimmed)) {
            return { 
                valid: false, 
                error: 'forbidden_chars',
                forbiddenChars: this.FORBIDDEN_CHARS_DISPLAY
            };
        }
        
        // Check for reserved Windows names (case-insensitive)
        const reservedNames = /^(CON|PRN|AUX|NUL|COM[1-9]|LPT[1-9])(\.|$)/i;
        if (reservedNames.test(trimmed)) {
            return { valid: false, error: 'reserved_name' };
        }
        
        // Check for names starting/ending with dots or spaces (problematic on some systems)
        if (trimmed.startsWith('.') && trimmed.length === 1) {
            return { valid: false, error: 'invalid_dot' };
        }
        if (trimmed.endsWith('.') || trimmed.endsWith(' ')) {
            return { valid: false, error: 'trailing_dot_space' };
        }
        
        return { valid: true, sanitized: trimmed };
    },
    
    /**
     * Validate a path (may contain forward slashes for folder separators)
     * @param {string} path - The path to validate
     * @returns {{ valid: boolean, error?: string, sanitized?: string }}
     */
    validatePath(path) {
        if (!path || typeof path !== 'string') {
            return { valid: false, error: 'empty' };
        }
        
        const trimmed = path.trim();
        if (!trimmed) {
            return { valid: false, error: 'empty' };
        }
        
        // Split by forward slash and validate each segment
        const segments = trimmed.split('/').filter(s => s.length > 0);
        if (segments.length === 0) {
            return { valid: false, error: 'empty' };
        }
        
        for (const segment of segments) {
            const result = this.validateFilename(segment);
            if (!result.valid) {
                return result;
            }
        }
        
        // Rebuild path without empty segments
        return { valid: true, sanitized: segments.join('/') };
    }
};

function noteApp() {
    return {
        // ============================================
        // GROUPED STATE - 状态分组结构
        // ============================================
        
        // App 全局状态
        app: {
            name: 'GoNote',
            version: '0.0.0',
            authEnabled: false,
            demoMode: false,
            alreadyDonated: true,
        },
        
        // 当前笔记状态
        note: {
            current: '',           // 当前笔记路径
            name: '',              // 当前笔记名称
            content: '',           // 当前笔记内容
            metadata: null,        // frontmatter 元数据
            isSaving: false,       // 保存中
            lastSaved: false,      // 已保存标记
            saveTimeout: null,     // 保存防抖定时器
            outline: [],           // 目录大纲
            backlinks: [],         // 反向链接
            backlinksLoading: false, // 反向链接加载中
        },
        
        // 笔记列表
        notes: [],
        
        // 每文件夹笔记分页状态（用于侧边栏渲染虚拟化）
        folderNotePages: {},
        
        // 搜索状态
        search: {
            query: '',
            mode: 'full',          // 'full' | 'title' | 'smart'
            results: [],
            page: 1,
            limit: 20,
            totalPages: 1,
            totalResults: 0,
            isSearching: false,
            debounceTimeout: null,
            highlight: '',         // 当前高亮搜索词
            matchIndex: 0,
            totalMatches: 0,
            editorMatchPositions: [],
            editorMatchCount: 0,
            currentEditorMatchIndex: -1,
            targetLineNumber: 0,   // 搜索结果目标行号
        },
        
        // 图谱状态
        graph: {
            show: false,
            instance: null,
            loaded: false,
            data: null,
            lastTheme: null,       // Mermaid 主题缓存
        },
        
        // 主题状态
        theme: {
            current: 'light',
            available: [],
        },
        
        // 国际化状态
        i18n: {
            locale: (function() {
                const saved = localStorage.getItem('locale');
                if (saved) return saved;
                const browserLang = navigator.language || navigator.userLanguage || 'zh-CN';
                if (browserLang.startsWith('zh')) return 'zh-CN';
                if (browserLang.startsWith('en')) return 'en-US';
                return 'zh-CN';
            })(),
            available: [],
            translations: window.__preloadedTranslations || {},
        },
        
        // 文件夹状态
        folders: {
            tree: [],
            all: [],
            expanded: new Set(),
            dragOver: null,
        },
        
        // 标签状态
        tags: {
            all: {},
            selected: [],
            expanded: false,
            reloadTimeout: null,
            // Lazy loading state for backend filtering
            notes: [],           // Filtered notes from backend
            page: 1,             // Current page
            limit: 50,           // Notes per page
            hasMore: false,      // Has more pages
            loading: false,      // Loading state
            total: 0,            // Total matching notes
        },
        
        // UI 状态
        ui: {
            viewMode: 'split',     // 'edit', 'split', 'preview'
            zenMode: false,
            previousViewMode: 'split',
            activePanel: 'files',
            sidebarWidth: CONFIG.DEFAULT_SIDEBAR_WIDTH,
            editorWidth: 50,
            isResizing: false,
            isResizingSplit: false,
            mobileSidebarOpen: false,
            readableLineLength: true,
            syntaxHighlightEnabled: false,
            syntaxHighlightTimeout: null,
            hideUnderscoreFolders: localStorage.getItem('hideUnderscoreFolders') === 'true',
            isScrolling: false,
            linkCopied: false,
        },
        
        // WebSocket 状态
        ws: {
            connection: null,
            reconnectTimeout: null,
            reconnectAttempts: 0,
            lastSyncTimestamp: null,
            disabled: false,  // Set to true when max reconnect attempts reached
        },
        
        // 模态框状态
        modals: {
            // 新建下拉菜单
            newDropdown: {
                show: false,
                targetFolder: null,
                position: { top: 0, left: 0 },
            },
            // 模板
            template: {
                show: false,
                available: [],
                selected: '',
                newNoteName: '',
            },
            // 分享
            share: {
                show: false,
                info: null,
                loading: false,
                showQR: false,
                linkCopied: false,
            },
            // 快速切换
            quickSwitcher: {
                show: false,
                query: '',
                index: 0,
                results: [],
            },
            // 孤立媒体清理
            orphanedMedia: {
                show: false,
                scanned: false,
                files: [],
                loading: false,
                error: null,
                totalSize: 0,
                cleanupInProgress: false,
                cleanupSuccess: false,
            },
            // 确认对话框（替换原生 confirm）
            confirm: {
                show: false,
                message: '',
                resolve: null,
            },
            // 提示对话框（替换原生 alert）
            alert: {
                show: false,
                message: '',
            },
            // 输入对话框（替换原生 prompt）
            prompt: {
                show: false,
                message: '',
                value: '',
                resolve: null,
            },
        },
        
        // 拖拽状态
        drag: {
            item: null,            // { path, type }
            target: null,          // 'editor' | 'folder' | null
        },
        
        // 撤销/重做历史
        history: {
            undo: [],
            redo: [],
            maxSize: CONFIG.MAX_UNDO_HISTORY,
            isUndoRedo: false,
            hasPendingChanges: false,
        },
        
        // 收藏状态（内部存储，通过 getter 访问）
        _favoritesState: {
            list: [],
            set: new Set(),
            expanded: true,
        },
        
        // 统计插件状态
        stats: {
            enabled: true,
            data: null,
            expanded: false,
        },
        
        // 媒体状态
        media: {
            current: '',
            type: 'image',         // 'image', 'audio', 'video', 'document'
            attachments: [],
            attachmentsLoading: false,
            attachmentsExpanded: false,
        },
        
        // 首页状态
        homepage: {
            selectedFolder: '',
            maxNotes: 50,
        },
        
        // ============================================
        // COMPUTED HELPERS - 计算属性
        // ============================================
        
        // Computed-like helpers for homepage (cached for performance)
        homepageNotes() {
            // Return cached result if folder hasn't changed
            if (NoteCache.homepageCache.folderPath === this.homepage.selectedFolder && NoteCache.homepageCache.notes) {
                return NoteCache.homepageCache.notes;
            }
            
            if (!this.folders.tree || typeof this.folders.tree !== 'object') {
                return [];
            }
            
            const folderNode = this.getFolderNode(this.homepage.selectedFolder || '');
            const result = (folderNode && Array.isArray(folderNode.notes)) ? folderNode.notes : [];
            
            // Cache the result
            NoteCache.homepageCache.notes = result;
            NoteCache.homepageCache.folderPath = this.homepage.selectedFolder;
            
            return result;
        },
        
        homepageFolders() {
            // Return cached result if folder hasn't changed
            if (NoteCache.homepageCache.folderPath === this.homepage.selectedFolder && NoteCache.homepageCache.folders) {
                return NoteCache.homepageCache.folders;
            }
            
            if (!this.folders.tree || typeof this.folders.tree !== 'object') {
                return [];
            }
            
            // Get child folders
            let childFolders = [];
            if (!this.homepage.selectedFolder) {
                // Root level: all top-level folders
                childFolders = Object.entries(this.folders.tree)
                    .filter(([key]) => key !== '__root__')
                    .map(([, folder]) => folder);
            } else {
                // Inside a folder: get its children
                const parentFolder = this.getFolderNode(this.homepage.selectedFolder);
                if (parentFolder && parentFolder.children) {
                    childFolders = Object.values(parentFolder.children);
                }
            }
            
            // Map to simplified structure (note count already cached in folder node)
            const result = childFolders
                .map(folder => ({
                    name: folder.name,
                    path: folder.path,
                    noteCount: folder.noteCount || 0  // Use pre-calculated count
                }))
                .sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()));
            
            // Cache the result
            NoteCache.homepageCache.folders = result;
            NoteCache.homepageCache.folderPath = this.homepage.selectedFolder;
            
            return result;
        },
        
        homepageBreadcrumb() {
            // Return cached result if folder hasn't changed
            if (NoteCache.homepageCache.folderPath === this.homepage.selectedFolder && NoteCache.homepageCache.breadcrumb) {
                return NoteCache.homepageCache.breadcrumb;
            }
            
            const breadcrumb = [{ name: this.t('homepage.title'), path: '' }];
            
            if (this.homepage.selectedFolder) {
                const parts = this.homepage.selectedFolder.split('/').filter(Boolean);
                let currentPath = '';
                
                parts.forEach(part => {
                    currentPath = currentPath ? `${currentPath}/${part}` : part;
                    breadcrumb.push({ name: part, path: currentPath });
                });
            }
            
            // Cache the result
            NoteCache.homepageCache.breadcrumb = breadcrumb;
            NoteCache.homepageCache.folderPath = this.homepage.selectedFolder;
            
            return breadcrumb;
        },
        
        // Helper: Format file size nicely
        formatSize(bytes) {
            if (!bytes) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
        },
        
        formatDate(dateStr) {
            if (!dateStr) return '';
            const date = new Date(dateStr);
            if (isNaN(date.getTime())) return '';
            return date.toLocaleDateString(this.i18n.locale, { 
                year: 'numeric', 
                month: 'short', 
                day: 'numeric' 
            });
        },
        
        getFolderNode(folderPath = '') {
            if (!this.folders.tree || typeof this.folders.tree !== 'object') {
                return null;
            }
            
            if (!folderPath) {
                return this.folders.tree['__root__'] || { name: '', path: '', children: {}, notes: [], noteCount: 0 };
            }
            
            const parts = folderPath.split('/').filter(Boolean);
            let currentLevel = this.folders.tree;
            let node = null;
            
            for (const part of parts) {
                if (!currentLevel[part]) {
                    return null;
                }
                node = currentLevel[part];
                currentLevel = node.children || {};
            }
            
            return node;
        },
        
        // Check if app is empty (no notes and no folders)
        get isAppEmpty() {
            const notesArray = Array.isArray(this.notes) ? this.notes : [];
            const foldersArray = Array.isArray(this.folders.all) ? this.folders.all : [];
            return notesArray.length === 0 && foldersArray.length === 0;
        }, 
        
        // Initialize app
        async init() {
            // Prevent double initialization (Alpine.js may call x-init twice in some cases)
            if (window.__noteapp_initialized) return;
            window.__noteapp_initialized = true;
            
            try {
                // Store global reference for native event handlers in x-html content
                window.$root = this;
                // Expose modal helpers for ErrorHandler and x-html event handlers
                window.showAppAlert = this.showAlert.bind(this);
                
                // ESC key to cancel drag operations
                document.addEventListener('keydown', (e) => {
                    if (e.key === 'Escape' && this.drag.item) {
                        this.cancelDrag();
                    }
                });
                
                await this.loadConfig();
                await this.loadThemes();
                await this.initTheme();
                await this.loadAvailableLocales();
                // Note: Translations are preloaded synchronously before Alpine init (see index.html)
                // loadLocale() is only called when user changes language from settings
                await this.loadNotes();
                await this.loadSharedNotePaths();
                await this.loadTemplates();
                this.loadLocalSettings();
                
                // Check for unsaved drafts from previous session
                this.checkAndRestoreDrafts();
                
                // Warn on page close if there are unsaved changes
                window.addEventListener('beforeunload', (e) => {
                    if (this.note && this.note.dirty) {
                        e.preventDefault();
                        e.returnValue = '';
                    }
                });
                
                // Periodic draft save while editing
                this._draftTimer = setInterval(() => {
                    if (this.note && this.note.dirty && this.note.current && this.note.content) {
                        this.writeDraft(this.note.current, this.note.content);
                    }
                }, 2000);
                
                // Parse URL and load specific note if provided
                this.loadItemFromURL();
                
                // Set initial homepage state ONLY if we're actually on the homepage
                if (window.location.pathname === '/') {
                    window.history.replaceState({ homepageFolder: '' }, '', '/');
                    document.title = this.app.name;
                }
                
                // Listen for browser back/forward navigation
                window.addEventListener('popstate', (e) => {
                    if (e.state && e.state.notePath) {
                        // Navigating to a note
                        const searchQuery = e.state.searchQuery || '';
                        this.loadNote(e.state.notePath, false, searchQuery); // false = don't update history
                        
                        // Update search box and trigger search if needed
                        if (searchQuery) {
                            this.search.query = searchQuery;
                        this.searchNotes();
                    } else {
                        this.search.query = '';
                        this.search.results = [];
                        this.clearSearchHighlights();
                    }
                } else if (e.state && e.state.mediaPath) {
                    // Navigating to a media file
                    this.viewMedia(e.state.mediaPath, null, false);
                } else {
                    // Navigating back to homepage
                    this.note.current = '';
                    this.note.content = '';
                    this.note.name = '';
                    this.note.outline = [];
                    this.modals.share.info = null; // Reset share info
                    document.title = this.app.name;
                    
                    // Restore homepage folder state if it was saved
                    if (e.state && e.state.homepageFolder !== undefined) {
                        this.homepage.selectedFolder = e.state.homepageFolder || '';
                    } else {
                        // No folder state in history, go to root
                        this.homepage.selectedFolder = '';
                    }
                    
                    // Invalidate cache to force recalculation
                    NoteCache.resetHomepageCache();
                    
                    // Clear search
                    this.search.query = '';
                    this.search.results = [];
                    this.clearSearchHighlights();
                }
            });
            
            // Cache DOM references after initial render
            this.$nextTick(() => {
                this.refreshDOMCache();
            });
            
            // Setup mobile view mode handler
            this.setupMobileViewMode();
            
            // Watch view mode changes and auto-save
            this.$watch('ui.viewMode', (newValue) => {
                this.saveViewMode();
                // Scroll to top when switching modes
                this.$nextTick(() => {
                    this.scrollToTop();
                });
            });
            
            // Watch for changes in note content to re-apply search highlights
            this.$watch('note.content', () => {
                if (this.search.highlight) {
                    // Re-apply highlights after content changes (with small delay for render)
                    this.$nextTick(() => {
                        setTimeout(() => {
                            // Don't focus editor during content changes (false)
                            this.highlightSearchTerm(this.search.highlight, false);
                        }, 50);
                    });
                }
            });
            
            // Watch tags panel expanded state and save to localStorage
            this.$watch('tags.expanded', () => {
                this.saveTagsExpanded();
            });
            
            // Watch favorites expanded state and save to localStorage
            this.$watch('_favoritesState.expanded', () => {
                this.saveFavoritesExpanded();
            });
            
            // Setup keyboard shortcuts (only once to prevent double triggers)
            if (!window.__noteapp_shortcuts_initialized) {
                window.__noteapp_shortcuts_initialized = true;
                window.addEventListener('keydown', (e) => {
                    // Use e.key (not e.code) for letter keys to support non-QWERTY keyboard layouts
                    
                    // Ctrl/Cmd + S to save
                    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
                        e.preventDefault();
                        this.saveNote();
                    }
                    
                    // Ctrl/Cmd + Alt + P for Quick Switcher
                    if ((e.ctrlKey || e.metaKey) && e.altKey && e.key.toLowerCase() === 'p') {
                        e.preventDefault();
                        this.openQuickSwitcher();
                        return;
                    }
                    
                    // Ctrl/Cmd + Alt/Option + N for new note
                    if ((e.ctrlKey || e.metaKey) && e.altKey && e.key.toLowerCase() === 'n') {
                        e.preventDefault();
                        this.createNote();
                    }
                    
                    // Ctrl/Cmd + Alt/Option + F for new folder
                    if ((e.ctrlKey || e.metaKey) && e.altKey && e.key.toLowerCase() === 'f') {
                        e.preventDefault();
                        this.createFolder();
                    }
                    
                    // Ctrl/Cmd + Z for undo (without shift or alt)
                    // Use e.key instead of e.code to support non-QWERTY keyboard layouts
                    if ((e.ctrlKey || e.metaKey) && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'z') {
                        e.preventDefault();
                        this.undo();
                    }
                    
                    // Ctrl/Cmd + Y OR Ctrl/Cmd+Shift+Z for redo
                    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'y') {
                        e.preventDefault();
                        this.redo();
                    }
                    if ((e.ctrlKey || e.metaKey) && e.shiftKey && !e.altKey && e.key.toLowerCase() === 'z') {
                        e.preventDefault();
                        this.redo();
                    }
                    
                    // F3 for next search match
                    if (e.code === 'F3' && !e.shiftKey) {
                        e.preventDefault();
                        this.nextMatch();
                    }
                    
                    // Shift + F3 for previous search match
                    if (e.code === 'F3' && e.shiftKey) {
                        e.preventDefault();
                        this.previousMatch();
                    }
                    
                    // Only apply markdown shortcuts when editor is focused and a note is open
                    const isEditorFocused = document.activeElement?.id === 'note-editor';
                    if (isEditorFocused && this.note.current) {
                        // Ctrl/Cmd + B for bold
                        if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'b') {
                            e.preventDefault();
                            this.wrapSelection('**', '**', 'bold text');
                        }
                        
                        // Ctrl/Cmd + I for italic
                        if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'i') {
                            e.preventDefault();
                            this.wrapSelection('*', '*', 'italic text');
                        }
                        
                        // Ctrl/Cmd + K for link
                        if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
                            e.preventDefault();
                            this.insertLink();
                        }
                        
                        // Ctrl/Cmd + Alt/Option + T for table
                        if ((e.ctrlKey || e.metaKey) && e.altKey && e.key.toLowerCase() === 't') {
                            e.preventDefault();
                            this.insertTable();
                        }
                        
                        // Ctrl/Cmd + Alt/Option + Z for Zen mode
                        if ((e.ctrlKey || e.metaKey) && e.altKey && e.key.toLowerCase() === 'z') {
                            e.preventDefault();
                            this.toggleZenMode();
                        }
                    }
                    
                    // Escape to exit Zen mode (works anywhere)
                    if (e.key === 'Escape' && this.ui.zenMode) {
                        e.preventDefault();
                        this.toggleZenMode();
                    }
                });
            }
            
            // Note: setupScrollSync() is called when a note is loaded (see loadNote())
            
            // Listen for system theme changes
            if (window.matchMedia) {
                window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
                    if (this.theme.current === 'system') {
                        this.applyTheme('system');
                    }
                });
            }
            
            // Listen for fullscreen changes (to sync zen mode state)
            document.addEventListener('fullscreenchange', () => {
                if (!document.fullscreenElement && this.ui.zenMode) {
                    // User exited fullscreen manually, exit zen mode too
                    this.ui.zenMode = false;
                    this.ui.viewMode = this.ui.previousViewMode;
                }
            });
            
            this.startPolling();
            
            document.addEventListener('visibilitychange', () => {
                if (document.hidden) {
                    this.stopPolling();
                } else {
                    this.startPolling();
                }
            });
            } catch (error) {
                console.error('App initialization failed:', error);
            }
        },
        
        // Load app configuration
        async loadConfig() {
            try {
                const response = await fetch('/api/config');
                const config = await response.json();
                this.app.name = config.name;
                this.app.version = config.version || '0.0.0';
                this.app.authEnabled = config.authentication?.enabled || false;
                this.app.demoMode = config.demoMode || false;
                this.app.alreadyDonated = config.alreadyDonated || true;
            } catch (error) {
                console.error('Failed to load config:', error);
            }
        },
        
        // Logout - sends POST request with CSRF protection
        async logout() {
            try {
                const response = await secureFetch('/logout', {
                    method: 'POST'
                });
                
                if (response.ok) {
                    // Clear local state
                    this.note.current = '';
                    this.note.content = '';
                    this.note.name = '';
                    // Redirect to login page
                    window.location.href = '/login';
                } else {
                    const error = await response.json().catch(() => ({}));
                    this.showAlert(error.detail || this.t('settings.logout_failed') || 'Logout failed');
                }
            } catch (error) {
                console.error('Logout failed:', error);
                this.showAlert(this.t('settings.logout_failed') || 'Logout failed');
            }
        },
        
        // Load available themes from backend
        async loadThemes() {
            try {
                const response = await fetch('/api/themes');
                const data = await response.json();
                
                // Use theme names directly from backend (already include emojis)
                this.theme.available = data.themes;
            } catch (error) {
                console.error('Failed to load themes:', error);
                // Fallback to default themes
                this.theme.available = [
                    { id: 'light', name: '🌞 Light' },
                    { id: 'dark', name: '🌙 Dark' }
                ];
            }
        },
        
        // Initialize theme system
        async initTheme() {
            // Load saved theme preference from localStorage
            const savedTheme = localStorage.getItem('gonoteTheme') || 'light';
            this.theme.current = savedTheme;
            await this.applyTheme(savedTheme);
        },
        
        // Set and apply theme
        async setTheme(themeId) {
            this.theme.current = themeId;
            localStorage.setItem('gonoteTheme', themeId);
            await this.applyTheme(themeId);
        },
        
        // Syntax highlighting toggle
        toggleSyntaxHighlight() {
            this.ui.syntaxHighlightEnabled = !this.ui.syntaxHighlightEnabled;
            localStorage.setItem('syntaxHighlightEnabled', this.ui.syntaxHighlightEnabled);
            if (this.ui.syntaxHighlightEnabled) {
                this.updateSyntaxHighlight();
            }
        },
        
        // Load all localStorage settings at once using centralized config
        loadLocalSettings() {
            for (const [prop, config] of Object.entries(LOCAL_SETTINGS)) {
                try {
                    const saved = localStorage.getItem(config.key);
                    let value;
                    
                    if (saved === null) {
                        // Use default value if not set
                        value = config.default;
                    } else if (config.type === 'boolean') {
                        value = saved === 'true';
                    } else if (config.type === 'number') {
                        const num = parseFloat(saved);
                        // Validate range if specified
                        if (!isNaN(num) && 
                            (config.min === undefined || num >= config.min) && 
                            (config.max === undefined || num <= config.max)) {
                            value = num;
                        } else {
                            value = config.default;
                        }
                    } else if (config.type === 'string') {
                        // Validate against allowed values if specified
                        if (!config.valid || config.valid.includes(saved)) {
                            value = saved;
                        } else {
                            value = config.default;
                        }
                    } else if (config.type === 'json') {
                        value = JSON.parse(saved);
                    }
                    
                    // Set value to the correct target location using dot notation
                    if (config.target) {
                        const parts = config.target.split('.');
                        let obj = this;
                        for (let i = 0; i < parts.length - 1; i++) {
                            obj = obj[parts[i]];
                        }
                        obj[parts[parts.length - 1]] = value;
                    }
                } catch (error) {
                    console.error(`Error loading setting ${prop}:`, error);
                    // Set default value to target on error
                    if (config.target) {
                        const parts = config.target.split('.');
                        let obj = this;
                        for (let i = 0; i < parts.length - 1; i++) {
                            obj = obj[parts[i]];
                        }
                        obj[parts[parts.length - 1]] = config.default;
                    }
                }
            }
            
            // Special case: favorites also needs to update the Set for O(1) lookups
            this._favoritesState.set = new Set(this._favoritesState.list);
            
            // Special case: expandedFolders needs to convert array back to Set
            if (Array.isArray(this.folders.expanded)) {
                this.folders.expanded = new Set(this.folders.expanded);
            }
        },
        
        // Readable line length toggle (for preview max-width)
        toggleReadableLineLength() {
            this.ui.readableLineLength = !this.ui.readableLineLength;
            localStorage.setItem('readableLineLength', this.ui.readableLineLength);
        },
        
        // Hide underscore folders toggle (hides _attachments, _templates, etc. from sidebar)
        toggleHideUnderscoreFolders() {
            this.ui.hideUnderscoreFolders = !this.ui.hideUnderscoreFolders;
            localStorage.setItem('hideUnderscoreFolders', this.ui.hideUnderscoreFolders);
        },
        
        // Update syntax highlight overlay (debounced, called on input)
        updateSyntaxHighlight() {
            if (!this.ui.syntaxHighlightEnabled) return;
            
            clearTimeout(this.ui.syntaxHighlightTimeout);
            this.ui.syntaxHighlightTimeout = setTimeout(() => {
                const overlay = document.getElementById('syntax-overlay');
                if (overlay) {
                    overlay.innerHTML = this.highlightMarkdown(this.note.content);
                }
            }, 50); // 50ms debounce
        },
        
        // Sync overlay scroll with textarea
        syncOverlayScroll() {
            const textarea = document.getElementById('note-editor');
            const overlay = document.getElementById('syntax-overlay');
            if (textarea && overlay) {
                requestAnimationFrame(() => {
                    overlay.scrollTop = textarea.scrollTop;
                    overlay.scrollLeft = textarea.scrollLeft;
                });
            }
        },
        
        // Highlight markdown syntax
        highlightMarkdown(text) {
            if (!text) return '';
            
            // Escape HTML first
            let html = this.escapeHtml(text);
            
            // Store code blocks and inline code with placeholders to protect from other patterns
            const codePlaceholders = [];
            
            // Code blocks FIRST - protect them before anything else
            html = html.replace(/(```[\s\S]*?```)/g, (match) => {
                codePlaceholders.push('<span class="md-codeblock">' + match + '</span>');
                return `\x00CODE${codePlaceholders.length - 1}\x00`;
            });
            
            // Frontmatter (must be at VERY start of document, not any line)
            if (html.startsWith('---\n')) {
                html = html.replace(/^(---\n[\s\S]*?\n---)/, (match) => {
                    codePlaceholders.push('<span class="md-frontmatter">' + match + '</span>');
                    return `\x00CODE${codePlaceholders.length - 1}\x00`;
                });
            }
            
            // Inline code - protect it
            html = html.replace(/`([^`\n]+)`/g, (match) => {
                codePlaceholders.push('<span class="md-code">' + match + '</span>');
                return `\x00CODE${codePlaceholders.length - 1}\x00`;
            });
            
            // Now apply other patterns (they won't match inside protected code)
            
            // Headings - capture the whitespace to preserve exact characters (tabs vs spaces)
            // This prevents cursor/selection misalignment
            html = html.replace(/^(#{1,6})(\s)(.*)$/gm, '<span class="md-heading">$1$2$3</span>');
            
            // Bold (must come before italic)
            html = html.replace(/\*\*([^*]+)\*\*/g, '<span class="md-bold">**$1**</span>');
            html = html.replace(/__([^_]+)__/g, '<span class="md-bold">__$1__</span>');
            
            // Italic
            html = html.replace(/(?<![*\\])\*([^*\n]+)\*(?!\*)/g, '<span class="md-italic">*$1*</span>');
            html = html.replace(/(?<![_\\])_([^_\n]+)_(?!_)/g, '<span class="md-italic">_$1_</span>');
            
            // Wikilinks [[...]]
            html = html.replace(/\[\[([^\]]+)\]\]/g, '<span class="md-wikilink">[[$1]]</span>');
            
            // Links [text](url)
            html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<span class="md-link">[$1]</span><span class="md-link-url">($2)</span>');
            
            // Lists - use ([ \t]) to capture the space/tab and preserve exact characters
            // IMPORTANT: Don't add any characters (like \u200B) that aren't in the original,
            // as this breaks cursor/selection alignment between textarea and overlay
            html = html.replace(/^(\s*)([-*+])([ \t])(.*)$/gm, (match, indent, bullet, space, rest) => {
                return `${indent}<span class="md-list">${bullet}</span>${space}${rest}`;
            });
            html = html.replace(/^(\s*)(\d+\.)([ \t])(.*)$/gm, (match, indent, bullet, space, rest) => {
                return `${indent}<span class="md-list">${bullet}</span>${space}${rest}`;
            });
            
            // Blockquotes
            html = html.replace(/^(&gt;.*)$/gm, '<span class="md-blockquote">$1</span>');
            
            // Horizontal rules
            html = html.replace(/^([-*_]{3,})$/gm, '<span class="md-hr">$1</span>');
            
            // Restore protected code blocks
            html = html.replace(/\x00CODE(\d+)\x00/g, (match, index) => codePlaceholders[parseInt(index)]);
            
            // Add trailing space to match textarea's phantom line for cursor
            // This ensures the overlay and textarea have the same content height
            html += '\n ';
            
            return html;
        },
        
        // Apply theme to document
        async applyTheme(themeId) {
            // Load theme CSS from file
            try {
                const response = await fetch(`/api/themes/${themeId}`);
                const data = await response.json();
                
                // Create or update style element
                let styleEl = document.getElementById('dynamic-theme');
                if (!styleEl) {
                    styleEl = document.createElement('style');
                    styleEl.id = 'dynamic-theme';
                    document.head.appendChild(styleEl);
                }
                styleEl.textContent = data.css;
                
                // Set data attribute for theme-specific selectors
                document.documentElement.setAttribute('data-theme', themeId);
                
                // Load appropriate Highlight.js theme for code syntax highlighting
                const highlightTheme = document.getElementById('highlight-theme');
                if (highlightTheme) {
                    if (themeId === 'light') {
                        highlightTheme.href = '/static/libs/highlight.js/11.11.1/styles/github.min.css';
                    } else {
                        // Use dark theme for dark/custom themes
                        highlightTheme.href = '/static/libs/highlight.js/11.11.1/styles/github-dark.min.css';
                    }
                }
                
                // Re-render Mermaid diagrams with new theme if there's a current note
                if (this.note.current) {
                    // Small delay to allow theme CSS to load
                    setTimeout(() => {
                        // Clear existing Mermaid renders
                        const previewContent = NoteCache.domCache.previewContent || document.querySelector('.markdown-preview');
                        if (previewContent) {
                            const mermaidContainers = previewContent.querySelectorAll('.mermaid-rendered');
                            mermaidContainers.forEach(container => {
                                // Replace with the original code block for re-rendering
                                const parent = container.parentElement;
                                if (parent && container.dataset.originalCode) {
                                    const pre = document.createElement('pre');
                                    const code = document.createElement('code');
                                    code.className = 'language-mermaid';
                                    code.textContent = container.dataset.originalCode;
                                    pre.appendChild(code);
                                    parent.replaceChild(pre, container);
                                }
                            });
                        }
                        // Re-render with new theme
                        this.renderMermaid();
                    }, 100);
                }
                
                // Refresh graph if visible (longer delay to ensure CSS is applied)
                if (this.graph.show) {
                    setTimeout(() => this.initGraph(), 300);
                }
                
                // Update PWA theme-color meta tag to match current theme
                const themeColorMeta = NoteCache.domCache.themeColorMeta || document.querySelector('meta[name="theme-color"]');
                if (themeColorMeta) {
                    // Get the accent color from CSS variables
                    const accentColor = getComputedStyle(document.documentElement)
                        .getPropertyValue('--accent-primary').trim() || '#667eea';
                    themeColorMeta.setAttribute('content', accentColor);
                }
            } catch (error) {
                console.error('Failed to load theme:', error);
            }
        },
        
        // ==================== INTERNATIONALIZATION ====================
        
        // Translation function - get translated string by key
        t(key, params = {}) {
            const keys = key.split('.');
            let value = this.i18n.translations;
            
            for (const k of keys) {
                value = value?.[k];
            }
            
            // Fallback to key if translation not found (silently - default translations are inline)
            if (typeof value !== 'string') {
                return key;
            }
            
            // Replace {{param}} placeholders
            return value.replace(/\{\{(\w+)\}\}/g, (_, name) => params[name] ?? `{{${name}}}`);
        },
        
        /**
         * Get localized error message from FilenameValidator result
         * @param {object} validation - The validation result from FilenameValidator
         * @param {string} type - 'note' or 'folder'
         * @returns {string} Localized error message
         */
        getValidationErrorMessage(validation, type = 'note') {
            switch (validation.error) {
                case 'empty':
                    return type === 'note' 
                        ? this.t('notes.empty_name') 
                        : this.t('folders.invalid_name');
                case 'forbidden_chars':
                    return this.t('validation.forbidden_chars', { 
                        chars: validation.forbiddenChars 
                    });
                case 'reserved_name':
                    return this.t('validation.reserved_name');
                case 'invalid_dot':
                    return this.t('validation.invalid_dot');
                case 'trailing_dot_space':
                    return this.t('validation.trailing_dot_space');
                default:
                    return type === 'note' 
                        ? this.t('notes.invalid_name') 
                        : this.t('folders.invalid_name');
            }
        },
        
        // Load available locales from backend
        async loadAvailableLocales() {
            try {
                const response = await fetch('/api/locales');
                const data = await response.json();
                this.i18n.available = data.locales || [];
            } catch (error) {
                console.error('Failed to load available locales:', error);
                this.i18n.available = [{ code: 'en-US', name: 'English', flag: '🇺🇸' }];
            }
        },
        
        // Load translations for a specific locale
        async loadLocale(localeCode = null) {
            const targetLocale = localeCode || localStorage.getItem('locale') || 'en-US';
            
            try {
                const response = await fetch(`/api/locales/${targetLocale}`);
                if (response.ok) {
                    this.i18n.translations = await response.json();
                    this.i18n.locale = targetLocale;
                    localStorage.setItem('locale', targetLocale);
                } else if (targetLocale !== 'en-US') {
                    // Fallback to en-US if requested locale not found
                    await this.loadLocale('en-US');
                }
            } catch (error) {
                console.error('Failed to load locale:', error);
                // If en-US also fails, translations will be empty and t() will return keys
                if (targetLocale !== 'en-US') {
                    await this.loadLocale('en-US');
                }
            }
        },
        
        // Change locale and reload translations
        async changeLocale(localeCode) {
            await this.loadLocale(localeCode);
        },
        
        // ==================== END INTERNATIONALIZATION ====================
        
        // Load all notes
        async loadNotes() {
            try {
                // Load all notes with a high limit to ensure tags filtering works correctly
                // The backend returns notes sorted by modified date (newest first),
                // but we need all notes for tag filtering to work properly
                const response = await fetch('/api/notes?include_media=true&limit=10000');
                const data = await response.json();
                this.notes = data.notes || [];
                this.folders.all = data.folders || [];
                this.folderNotePages = {}; // Reset sidebar pagination on data reload
                this.buildNoteLookupMaps(); // Build O(1) lookup maps
                this.buildFolderTree();
                await this.loadTags(); // Load tags after notes are loaded
            } catch (error) {
                ErrorHandler.handle('load notes', error);
            }
        },
        
        // Build lookup maps for O(1) wikilink resolution
        buildNoteLookupMaps() {
            // Clear existing maps
            NoteCache.clearNoteCaches();
            
            if (!this.notes || !Array.isArray(this.notes)) {
                return;
            }
            
            for (const note of this.notes) {
                const path = note.path;
                const pathLower = path.toLowerCase();
                const name = note.name;
                const nameLower = name.toLowerCase();
                
                // Handle media files separately - build media lookup map
                if (note.type !== 'note') {
                    // Map filename WITH extension (case-insensitive) to full path
                    // Use path to get filename with extension (note.name is stem without extension)
                    const filenameWithExt = path.split('/').pop().toLowerCase();
                    // First match wins if there are duplicates
                    if (!NoteCache.mediaLookup.has(filenameWithExt)) {
                        NoteCache.mediaLookup.set(filenameWithExt, path);
                    }
                    continue;
                }
                
                // Notes only from here
                const nameWithoutMd = name.replace(/\.md$/i, '');
                const nameWithoutMdLower = nameWithoutMd.toLowerCase();
                
                // Store all variations for fast lookup
                NoteCache.noteLookup.byPath.set(path, true);
                NoteCache.noteLookup.byPath.set(path.replace(/\.md$/i, ''), true);
                NoteCache.noteLookup.byPathLower.set(pathLower, true);
                NoteCache.noteLookup.byPathLower.set(pathLower.replace(/\.md$/i, ''), true);
                NoteCache.noteLookup.byName.set(name, true);
                NoteCache.noteLookup.byName.set(nameWithoutMd, true);
                NoteCache.noteLookup.byNameLower.set(nameLower, true);
                NoteCache.noteLookup.byNameLower.set(nameWithoutMdLower, true);
                
                // End path matching (for /folder/note style links)
                NoteCache.noteLookup.byEndPath.set('/' + nameWithoutMdLower, true);
                NoteCache.noteLookup.byEndPath.set('/' + nameLower, true);
            }
        },
        
        // Fast O(1) check if a wikilink target exists
        wikiLinkExists(linkTarget) {
            const targetLower = linkTarget.toLowerCase();
            
            // Check all lookup maps
            return (
                NoteCache.noteLookup.byPath.has(linkTarget) ||
                NoteCache.noteLookup.byPath.has(linkTarget + '.md') ||
                NoteCache.noteLookup.byPathLower.has(targetLower) ||
                NoteCache.noteLookup.byPathLower.has(targetLower + '.md') ||
                NoteCache.noteLookup.byName.has(linkTarget) ||
                NoteCache.noteLookup.byNameLower.has(targetLower) ||
                NoteCache.noteLookup.byEndPath.has('/' + targetLower) ||
                NoteCache.noteLookup.byEndPath.has('/' + targetLower + '.md')
            );
        },
        
        // Resolve media wikilink to full path (O(1) lookup)
        // Returns the full path if found, null otherwise
        resolveMediaWikilink(mediaName) {
            const nameLower = mediaName.toLowerCase();
            return NoteCache.mediaLookup.get(nameLower) || null;
        },
        
        // Load all tags
        async loadTags() {
            try {
                const response = await fetch('/api/tags');
                const data = await response.json();
                this.tags.all = data.tags || {};
            } catch (error) {
                ErrorHandler.handle('load tags', error, false); // Don't show alert, tags are optional
            }
        },
        
        // Debounced tag reload (prevents excessive API calls during typing)
        loadTagsDebounced() {
            // Clear existing timeout
            if (this.tags.reloadTimeout) {
                clearTimeout(this.tags.reloadTimeout);
            }
            
            // Set new timeout - reload tags 2 seconds after last save
            this.tags.reloadTimeout = setTimeout(() => {
                this.loadTags();
            }, 2000);
        },
        
        // Toggle tag selection for filtering
        toggleTag(tag) {
            const index = this.tags.selected.indexOf(tag);
            if (index > -1) {
                this.tags.selected.splice(index, 1);
            } else {
                this.tags.selected.push(tag);
            }
            
            // Reset pagination and load filtered notes from backend
            this.tags.page = 1;
            this.tags.notes = [];
            this.tags.hasMore = false;
            this.loadNotesByTags();
        },
        
        // Load notes filtered by selected tags from backend (with pagination)
        async loadNotesByTags() {
            // If no tags selected, clear and let applyFilters show full tree
            if (this.tags.selected.length === 0) {
                this.tags.notes = [];
                this.tags.page = 1;
                this.tags.hasMore = false;
                this.tags.total = 0;
                this.applyFilters();
                return;
            }
            
            this.tags.loading = true;
            
            try {
                const tagsParam = this.tags.selected.map(t => encodeURIComponent(t)).join(',');
                const response = await fetch(
                    `/api/notes?tags=${tagsParam}&page=${this.tags.page}&limit=${this.tags.limit}&include_media=true`
                );
                const data = await response.json();
                
                if (this.tags.page === 1) {
                    this.tags.notes = data.notes || [];
                } else {
                    // Append for infinite scroll
                    this.tags.notes = [...this.tags.notes, ...(data.notes || [])];
                }
                
                // Update pagination state
                if (data.pagination) {
                    this.tags.hasMore = data.pagination.has_next || false;
                    this.tags.total = data.pagination.total || 0;
                } else {
                    this.tags.hasMore = false;
                    this.tags.total = (data.notes || []).length;
                }
                
                // Trigger UI update
                this.applyFilters();
            } catch (error) {
                console.error('Failed to load notes by tags:', error);
                ErrorHandler.handle('load notes by tags', error, false);
            } finally {
                this.tags.loading = false;
            }
        },
        
        // Load more notes for infinite scroll (lazy loading)
        async loadMoreTagNotes() {
            if (!this.tags.hasMore || this.tags.loading || this.tags.selected.length === 0) {
                return;
            }
            
            this.tags.page++;
            await this.loadNotesByTags();
        },
        
        // Load more notes per folder (rendering virtualisation for sidebar)
        loadMoreFolderNotes(folderPath) {
            const pages = this.folderNotePages[folderPath] || 1;
            this.folderNotePages[folderPath] = pages + 1;
            // Re-render by re-running buildFolderTree - the tree is stored and re-renders
            // via Alpine reactivity after updating the folders.tree reference
            this.buildFolderTree();
        },
        
        // Clear tag selection
        clearTagSelection() {
            this.tags.selected = [];
            this.tags.notes = [];
            this.tags.page = 1;
            this.tags.hasMore = false;
            this.tags.total = 0;
            this.applyFilters();
        },
        
        // ========================================================================
        // Template Methods
        // ========================================================================
        
        // Load available templates from _templates folder
        async loadTemplates() {
            try {
                const response = await fetch('/api/templates');
                const data = await response.json();
                this.modals.template.available = data.templates || [];
            } catch (error) {
                ErrorHandler.handle('load templates', error, false); // Don't show alert, templates are optional
            }
        },
        
        // Create a new note from a template
        async createNoteFromTemplate() {
            if (!this.modals.template.selected || !this.modals.template.newNoteName.trim()) {
                return;
            }
            
            try {
                // Validate the note name
                const validation = FilenameValidator.validateFilename(this.modals.template.newNoteName);
                if (!validation.valid) {
                    this.showAlert(this.getValidationErrorMessage(validation, 'note'));
                    return;
                }
                
                // Determine the note path based on dropdown context
                let notePath = validation.sanitized;
                if (!notePath.endsWith('.md')) {
                    notePath += '.md';
                }
                
                // Determine target folder: use dropdown context if set, otherwise homepage folder
                let targetFolder;
                if (this.modals.newDropdown.targetFolder !== null && this.modals.newDropdown.targetFolder !== undefined) {
                    targetFolder = this.modals.newDropdown.targetFolder; // Can be '' for root or a folder path
                } else {
                    targetFolder = this.homepage.selectedFolder || '';
                }
                
                // If we have a target folder, create note in that folder
                if (targetFolder) {
                    notePath = `${targetFolder}/${notePath}`;
                }
                
                // CRITICAL: Check if note already exists
                const existingNote = this.notes.find(note => note.path === notePath);
                if (existingNote) {
                    this.showAlert(this.t('notes.already_exists', { name: validation.sanitized }));
                    return;
                }
                
                // Create note from template
                const response = await secureFetch('/api/templates/create-note', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        templateName: this.modals.template.selected,
                        notePath: notePath
                    })
                });
                
                if (!response.ok) {
                    const error = await response.json();
                    this.showAlert(error.detail || this.t('templates.create_failed'));
                    return;
                }
                
                const data = await response.json();
                
                // Close modal and reset state
                this.modals.template.show = false;
                this.modals.template.selected = '';
                this.modals.template.newNoteName = '';
                
                // Reload notes and open the new note
                await this.loadNotes();
                await this.loadNote(data.path);
                this.focusEditorForNewNote();
                
            } catch (error) {
                ErrorHandler.handle('create note from template', error);
            }
        },
        
        // Clear all tag filters
        clearTagFilters() {
            this.tags.selected = [];
            
            // Apply unified filtering
            this.applyFilters();
        },
        
        // ========================================================================
        // Outline (TOC) Methods
        // ========================================================================
        
        // Extract headings from markdown content for the outline
        extractOutline(content) {
            if (!content) {
                this.note.outline = [];
                return;
            }
            
            const headings = [];
            const lines = content.split('\n');
            const slugCounts = {}; // Track duplicate slugs
            
            // Skip frontmatter and code blocks
            let inFrontmatter = false;
            let inCodeBlock = false;
            
            for (let i = 0; i < lines.length; i++) {
                const line = lines[i];
                
                // Handle frontmatter
                if (i === 0 && line.trim() === '---') {
                    inFrontmatter = true;
                    continue;
                }
                if (inFrontmatter) {
                    if (line.trim() === '---') {
                        inFrontmatter = false;
                    }
                    continue;
                }
                
                // Handle fenced code blocks (``` or ~~~)
                if (line.trim().startsWith('```') || line.trim().startsWith('~~~')) {
                    inCodeBlock = !inCodeBlock;
                    continue;
                }
                if (inCodeBlock) {
                    continue;
                }
                
                // Match heading lines (# to ######)
                const match = line.match(/^(#{1,6})\s+(.+)$/);
                if (match) {
                    const level = match[1].length;
                    const text = match[2].trim();
                    
                    // Generate slug (GitHub-style)
                    let slug = text
                        .toLowerCase()
                        .replace(/[^\w\s-]/g, '') // Remove special chars
                        .replace(/\s+/g, '-')     // Spaces to dashes
                        .replace(/-+/g, '-');     // Multiple dashes to single
                    
                    // Handle duplicate slugs
                    if (slugCounts[slug] !== undefined) {
                        slugCounts[slug]++;
                        slug = `${slug}-${slugCounts[slug]}`;
                    } else {
                        slugCounts[slug] = 0;
                    }
                    
                    headings.push({
                        level,
                        text,
                        slug,
                        line: i + 1 // 1-indexed line number
                    });
                }
            }
            
            this.note.outline = headings;
        },
        
        // Scroll to a heading in the editor or preview
        scrollToHeading(heading) {
            if (this.ui.viewMode === 'preview' || this.ui.viewMode === 'split') {
                // In preview/split mode, scroll the preview pane
                const preview = NoteCache.domCache.previewContent || document.querySelector('.markdown-preview');
                if (preview) {
                    // Try to find heading by ID first (most reliable)
                    const headingById = document.getElementById(heading.slug);
                    if (headingById) {
                        headingById.scrollIntoView({ behavior: 'smooth', block: 'start' });
                        this.highlightHeading(headingById);
                        return;
                    }
                    
                    // Fallback: find heading element by text content
                    const headingElements = preview.querySelectorAll('h1, h2, h3, h4, h5, h6');
                    for (const el of headingElements) {
                        // Compare trimmed text content
                        if (el.textContent.trim() === heading.text) {
                            el.scrollIntoView({ behavior: 'smooth', block: 'start' });
                            this.highlightHeading(el);
                            return;
                        }
                    }
                }
            }

            if (this.ui.viewMode === 'edit' || this.ui.viewMode === 'split') {
                // In edit/split mode, scroll the editor to the line
                const textarea = NoteCache.domCache.editor || document.querySelector('.editor-textarea');
                if (textarea && heading.line) {
                    const lines = textarea.value.split('\n');
                    let charPos = 0;

                    // Calculate character position of the heading line
                    for (let i = 0; i < heading.line - 1 && i < lines.length; i++) {
                        charPos += lines[i].length + 1; // +1 for newline
                    }

                    // Set cursor position and scroll
                    textarea.focus();
                    textarea.setSelectionRange(charPos, charPos);

                    // Calculate scroll position (approximate)
                    const lineHeight = parseFloat(getComputedStyle(textarea).lineHeight) || 24;
                    const paddingTop = parseFloat(getComputedStyle(textarea).paddingTop) || 24;
                    const scrollTop = (heading.line - 1) * lineHeight + paddingTop - textarea.clientHeight / 3;
                    textarea.scrollTop = Math.max(0, scrollTop);
                }
            }
        },

        // Add a brief highlight effect to a heading element
        highlightHeading(el) {
            el.style.transition = 'background-color 0.3s';
            el.style.backgroundColor = 'var(--accent-light)';
            setTimeout(() => {
                el.style.backgroundColor = '';
            }, 1000);
        },

        // Find the nearest heading above the editor viewport top
        findNearestHeadingInEditor(editor) {
            if (!this.note.outline || this.note.outline.length === 0) return null;
            const lineHeight = parseFloat(getComputedStyle(editor).lineHeight) || 24;
            const paddingTop = parseFloat(getComputedStyle(editor).paddingTop) || 24;
            const topLine = Math.max(1, Math.round((editor.scrollTop - paddingTop) / lineHeight) + 1);
            let nearest = null;
            for (const h of this.note.outline) {
                if (h.line <= topLine) nearest = h;
                else break;
            }
            return nearest;
        },

        // Find the heading closest to the preview viewport top
        findNearestHeadingInPreview(previewContainer) {
            if (!this.note.outline || this.note.outline.length === 0) return null;
            const containerRect = previewContainer.getBoundingClientRect();
            let nearest = null;
            let minDist = Infinity;
            for (const h of this.note.outline) {
                const el = document.getElementById(h.slug);
                if (el && previewContainer.contains(el)) {
                    const rect = el.getBoundingClientRect();
                    const dist = Math.abs(rect.top - containerRect.top);
                    if (dist < minDist) {
                        minDist = dist;
                        nearest = h;
                    }
                }
            }
            return nearest;
        },

        // Unified filtering logic combining tags and text search
        async applyFilters() {
            const hasTextSearch = this.search.query.trim().length > 0;
            const hasTagFilter = this.tags.selected.length > 0;
            
            // Case 1: No filters at all → show full folder tree
            if (!hasTextSearch && !hasTagFilter) {
                this.search.isSearching = false;
                this.search.results = [];
                this.search.highlight = '';
                this.clearSearchHighlights();
                this.buildFolderTree();
                return;
            }
            
            // Case 2: Only tag filter → use backend-filtered notes
            if (hasTagFilter && !hasTextSearch) {
                this.search.isSearching = false;
                // Use notes from backend filtering (supports lazy loading)
                this.search.results = this.tags.notes.filter(note => note.type === 'note');
                this.search.highlight = '';
                this.clearSearchHighlights();
                return;
            }
            
            // Case 3: Text search (with or without tag filter)
            if (hasTextSearch) {
                this.search.isSearching = true;
                try {
                    // Use pagination parameters from state, pass search mode
                    const url = `/api/search?q=${encodeURIComponent(this.search.query)}&mode=${encodeURIComponent(this.search.mode)}&page=${this.search.page}&limit=${this.search.limit}`;
                    const response = await fetch(url);
                    const data = await response.json();
                    
                    // Handle both old format (direct array) and new format (with pagination)
                    let results = Array.isArray(data) ? data : (data.results || []);
                    
                    // Extract pagination info if available
                    if (data.pagination) {
                        this.search.totalPages = data.pagination.total_pages || 1;
                        this.search.totalResults = data.pagination.total_items || results.length;
                    } else {
                        // Old format: assume all results in one page
                        this.search.totalPages = 1;
                        this.search.totalResults = results.length;
                    }
                    
                    // Apply tag filtering to search results if tags are selected
                    if (hasTagFilter) {
                        results = results.filter(result => {
                            const note = this.notes.find(n => n.path === result.path);
                            return note ? this.noteMatchesTags(note) : false;
                        });
                    }
                    
                    this.search.results = results;
                    
                    // Highlight search term in current note if open
                    if (this.note.current && this.note.content) {
                        this.search.highlight = this.search.query;
                        this.$nextTick(() => {
                            this.highlightSearchTerm(this.search.query, false);
                        });
                    }
                } catch (error) {
                    console.error('Search failed:', error);
                    this.search.results = [];
                } finally {
                    this.search.isSearching = false;
                }
            }
        },
        
        // Check if a note matches selected tags (AND logic)
        noteMatchesTags(note) {
            if (this.tags.selected.length === 0) {
                return true; // No filter active
            }
            if (!note.tags || note.tags.length === 0) {
                return false; // Note has no tags but filter is active
            }
            // Check if note has ALL selected tags (AND logic)
            return this.tags.selected.every(tag => note.tags.includes(tag));
        },
        
        // Get all tags sorted by name
        get sortedTags() {
            return Object.entries(this.tags.all).sort((a, b) => a[0].localeCompare(b[0]));
        },
        
        // Get tags for current note
        get currentNoteTags() {
            if (!this.note.current) return [];
            const note = this.notes.find(n => n.path === this.note.current);
            return note && note.tags ? note.tags : [];
        },
        
        // ==================== FAVORITES ====================
        
        // Save favorites to localStorage
        saveFavorites() {
            try {
                localStorage.setItem('noteFavorites', JSON.stringify(this._favoritesState.list));
            } catch (e) {
                console.warn('Could not save favorites to localStorage');
            }
        },
        
        // Check if a note is favorited (O(1) lookup)
        isFavorite(notePath) {
            return this._favoritesState.set.has(notePath);
        },
        
        // Toggle favorite status for a note
        toggleFavorite(notePath = null) {
            const path = notePath || this.note.current;
            if (!path) return;
            
            if (this._favoritesState.set.has(path)) {
                // Remove from favorites
                this._favoritesState.list = this._favoritesState.list.filter(f => f !== path);
            } else {
                // Add to favorites
                this._favoritesState.list = [...this._favoritesState.list, path];
            }
            // Recreate Set from array for consistency
            this._favoritesState.set = new Set(this._favoritesState.list);
            this.saveFavorites();
        },
        
        // Get favorite notes with full details (for display)
        get favoriteNotes() {
            return this._favoritesState.list
                .map(path => {
                    // Find note by exact path or case-insensitive match
                    let note = this.notes.find(n => n.path === path);
                    if (!note) {
                        note = this.notes.find(n => n.path.toLowerCase() === path.toLowerCase());
                    }
                    if (!note) return null;
                    return {
                        path: note.path, // Use actual path from notes (fixes case issues)
                        name: note.path.split('/').pop().replace('.md', ''),
                        folder: note.folder || ''
                    };
                })
                .filter(Boolean); // Remove nulls (deleted notes)
        },
        
        saveFavoritesExpanded() {
            try {
                localStorage.setItem('favoritesExpanded', this._favoritesState.expanded.toString());
            } catch (e) {
                console.error('Error saving favorites expanded state:', e);
            }
        },
        
        // Get current note's last modified time as relative string
        get lastEditedText() {
            if (!this.note.current) return '';
            const note = this.notes.find(n => n.path === this.note.current);
            if (!note || !note.modified) return '';
            
            const modified = new Date(note.modified);
            const now = new Date();
            const diffMs = now - modified;
            const diffSecs = Math.floor(diffMs / 1000);
            const diffMins = Math.floor(diffSecs / 60);
            const diffHours = Math.floor(diffMins / 60);
            const diffDays = Math.floor(diffHours / 24);
            
            if (diffSecs < 60) return this.t('editor.just_now');
            if (diffMins < 60) return this.t('editor.minutes_ago', { count: diffMins });
            if (diffHours < 24) return this.t('editor.hours_ago', { count: diffHours });
            if (diffDays < 7) return this.t('editor.days_ago', { count: diffDays });
            
            // For older dates, show the date in selected locale
            return modified.toLocaleDateString(this.i18n.locale, { month: 'short', day: 'numeric' });
        },
        
        // Parse tags from markdown content (matches backend logic)
        parseTagsFromContent(content) {
            if (!content || !content.trim().startsWith('---')) {
                return [];
            }
            
            try {
                const lines = content.split('\n');
                if (lines[0].trim() !== '---') return [];
                
                // Find closing ---
                let endIdx = -1;
                for (let i = 1; i < lines.length; i++) {
                    if (lines[i].trim() === '---') {
                        endIdx = i;
                        break;
                    }
                }
                
                if (endIdx === -1) return [];
                
                const frontmatterLines = lines.slice(1, endIdx);
                const tags = [];
                let inTagsList = false;
                
                for (const line of frontmatterLines) {
                    const stripped = line.trim();
                    
                    // Check for inline array: tags: [tag1, tag2]
                    if (stripped.startsWith('tags:')) {
                        const rest = stripped.substring(5).trim();
                        if (rest.startsWith('[') && rest.endsWith(']')) {
                            const tagsStr = rest.substring(1, rest.length - 1);
                            const rawTags = tagsStr.split(',').map(t => t.trim());
                            tags.push(...rawTags.filter(t => t).map(t => t.toLowerCase()));
                            break;
                        } else if (rest) {
                            tags.push(rest.toLowerCase());
                            break;
                        } else {
                            inTagsList = true;
                        }
                    } else if (inTagsList) {
                        if (stripped.startsWith('-')) {
                            const tag = stripped.substring(1).trim();
                            if (tag && !tag.startsWith('#')) {
                                tags.push(tag.toLowerCase());
                            }
                        } else if (stripped && !stripped.startsWith('#')) {
                            break;
                        }
                    }
                }
                
                return [...new Set(tags)].sort();
            } catch (e) {
                console.error('Error parsing tags:', e);
                return [];
            }
        },
        
        // Build folder tree structure
        buildFolderTree() {
            const tree = {};
            
            // Add ALL folders from backend (including empty ones)
            this.folders.all.forEach(folderPath => {
                const parts = folderPath.split('/');
                let current = tree;
                
                parts.forEach((part, index) => {
                    const fullPath = parts.slice(0, index + 1).join('/');
                    
                    if (!current[part]) {
                        current[part] = {
                            name: this.decodeNoteName(part),
                            path: fullPath,
                            children: {},
                            notes: []
                        };
                    }
                    current = current[part].children;
                });
            });
            
            // Add ALL notes to their folders (no filtering - tree only shown when no filters active)
            this.notes.forEach(note => {
                if (!note.folder) {
                    // Root level note
                    if (!tree['__root__']) {
                        tree['__root__'] = {
                            name: '',
                            path: '',
                            children: {},
                            notes: []
                        };
                    }
                    tree['__root__'].notes.push(note);
                } else {
                    // Navigate to the folder and add note
                    const parts = note.folder.split('/');
                    let current = tree;
                    
                    for (let i = 0; i < parts.length; i++) {
                        if (!current[parts[i]]) {
                            current[parts[i]] = {
                                name: parts[i],
                                path: parts.slice(0, i + 1).join('/'),
                                children: {},
                                notes: []
                            };
                        }
                        if (i === parts.length - 1) {
                            current[parts[i]].notes.push(note);
                        } else {
                            current = current[parts[i]].children;
                        }
                    }
                }
            });
            
            // Sort all notes arrays alphabetically (create new sorted arrays for reactivity)
            const sortNotes = (obj) => {
                if (obj.notes && obj.notes.length > 0) {
                    // Create a new sorted array instead of mutating for Alpine reactivity
                    obj.notes = [...obj.notes].sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()));
                }
                if (obj.children && Object.keys(obj.children).length > 0) {
                    Object.values(obj.children).forEach(child => sortNotes(child));
                }
            };
            
            // Sort notes in root (create new array for reactivity)
            if (tree['__root__'] && tree['__root__'].notes) {
                tree['__root__'].notes = [...tree['__root__'].notes].sort((a, b) => a.name.toLowerCase().localeCompare(b.name.toLowerCase()));
            }
            
            // Sort notes in all folders
            Object.values(tree).forEach(folder => {
                if (folder.path !== undefined) { // Skip __root__ as it was already sorted
                    sortNotes(folder);
                }
            });
            
            // Calculate and cache note counts recursively (for performance)
            const calculateNoteCounts = (folderNode) => {
                const directNotes = folderNode.notes ? folderNode.notes.length : 0;
                
                if (!folderNode.children || Object.keys(folderNode.children).length === 0) {
                    folderNode.noteCount = directNotes;
                    return directNotes;
                }
                
                const childNotesCount = Object.values(folderNode.children).reduce(
                    (total, child) => total + calculateNoteCounts(child),
                    0
                );
                
                folderNode.noteCount = directNotes + childNotesCount;
                return folderNode.noteCount;
            };
            
            // Calculate note counts for all folders
            Object.values(tree).forEach(folder => {
                if (folder.path !== undefined || folder === tree['__root__']) {
                    calculateNoteCounts(folder);
                }
            });
            
            // Invalidate homepage cache when tree is rebuilt
            NoteCache.resetHomepageCache();
            
            // Assign new tree (Alpine will detect the change)
            this.folders.tree = tree;
        },
        
        // =====================================================================
        // DATA-ATTRIBUTE BASED HANDLERS
        // These read path/name/type from data-* attributes, avoiding JS escaping issues
        // =====================================================================
        
        // Escape strings for HTML attributes (simpler than JS escaping)
        escapeHtmlAttr(str) {
            if (!str) return '';
            return str
                .replace(/&/g, '&amp;')
                .replace(/"/g, '&quot;')
                .replace(/'/g, '&#39;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;');
        },
        
        // Folder handlers - read from dataset
        handleFolderClick(el) {
            this.toggleFolder(el.dataset.path);
        },
        handleFolderDragOver(el, event) {
            event.preventDefault();
            this.folders.dragOver = el.dataset.path;
            el.classList.add('drag-over');
        },
        handleFolderDragLeave(el) {
            this.folders.dragOver = null;
            el.classList.remove('drag-over');
        },
        handleFolderDrop(el, event) {
            event.stopPropagation();
            el.classList.remove('drag-over');
            this.onFolderDrop(el.dataset.path);
        },
        handleNewItemClick(el, event) {
            event.stopPropagation();
            this.modals.newDropdown.targetFolder = el.dataset.path;
            this.toggleNewDropdown(event);
        },
        handleRenameFolderClick(el, event) {
            event.stopPropagation();
            this.renameFolder(el.dataset.path, el.dataset.name);
        },
        handleDeleteFolderClick(el, event) {
            event.stopPropagation();
            this.deleteFolder(el.dataset.path, el.dataset.name);
        },
        
        // Item (note/media) handlers - read from dataset
        handleItemClick(el) {
            this.openItem(el.dataset.path, el.dataset.type);
        },
        handleItemHover(el, isEnter) {
            const path = el.dataset.path;
            if (path !== this.note.current && path !== this.media.current) {
                el.style.backgroundColor = isEnter ? 'var(--bg-hover)' : 'transparent';
            }
        },
        handleDeleteItemClick(el, event) {
            event.stopPropagation();
            if (el.dataset.type === 'image') {
                this.deleteMedia(el.dataset.path);
            } else {
                this.deleteNote(el.dataset.path, el.dataset.name);
            }
        },
        
        // =====================================================================
        // FOLDER TREE RENDERING
        // =====================================================================
        
        // Render folder recursively (helper for deep nesting)
        // Uses data-* attributes to store path/name, avoiding JS string escaping issues
        renderFolderRecursive(folder, level = 0, isTopLevel = false) {
            if (!folder) return '';
            
            let html = '';
            const isExpanded = this.folders.expanded.has(folder.path);
            const esc = (s) => this.escapeHtmlAttr(s); // Shorthand for HTML escaping
            
            // Render this folder's header
            // Note: Using native event handlers with data-* attributes instead of Alpine directives
            // because x-html doesn't process Alpine directives in dynamically generated content
            html += `
                <div>
                    <div 
                        data-path="${esc(folder.path)}"
                        data-name="${esc(folder.name)}"
                        draggable="true"
                        ondragstart="window.$root.onItemDragStart(this.dataset.path, 'folder', event)"
                        ondragend="window.$root.onItemDragEnd()"
                        ondragover="window.$root.handleFolderDragOver(this, event)"
                        ondragenter="window.$root.handleFolderDragOver(this, event)"
                        ondragleave="window.$root.handleFolderDragLeave(this)"
                        ondrop="window.$root.handleFolderDrop(this, event)"
                        onclick="window.$root.handleFolderClick(this)"
                        class="folder-item hover-accent px-2 py-1 text-sm relative"
                        style="color: var(--text-primary); cursor: pointer;"
                    >
                        <div class="flex items-center gap-1">
                            <button 
                                class="flex-shrink-0 w-4 h-4 flex items-center justify-center"
                                style="color: var(--text-tertiary); cursor: pointer; transition: transform 0.2s; pointer-events: none; margin-left: -5px; ${isExpanded ? 'transform: rotate(90deg);' : ''}"
                            >
                                <svg width="12" height="12" viewBox="0 0 16 16" fill="currentColor">
                                    <path d="M6 4l4 4-4 4V4z"/>
                                </svg>
                            </button>
                            <span class="flex items-center gap-1 flex-1" style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-weight: 500; pointer-events: none;" title="${esc(folder.name)}">
                                <span>${esc(folder.name)}</span>
                                ${folder.notes.length === 0 && (!folder.children || Object.keys(folder.children).length === 0) ? `<span class="text-xs" style="color: var(--text-tertiary); font-weight: 400;">(${this.t('folders.empty')})</span>` : ''}
                            </span>
                        </div>
                        <div class="hover-buttons flex gap-1 transition-opacity absolute right-2 top-1/2 transform -translate-y-1/2" style="opacity: 0; pointer-events: none; background: linear-gradient(to right, transparent, var(--bg-hover) 20%, var(--bg-hover)); padding-left: 20px;" onclick="event.stopPropagation()">
                            <button 
                                data-path="${esc(folder.path)}"
                                onclick="window.$root.handleNewItemClick(this, event)"
                                class="px-1.5 py-0.5 text-xs rounded hover:brightness-110"
                                style="background-color: var(--bg-tertiary); color: var(--text-secondary);"
                                title="${this.t('folders.add_item')}"
                            >+</button>
                            <button 
                                data-path="${esc(folder.path)}"
                                data-name="${esc(folder.name)}"
                                onclick="window.$root.handleRenameFolderClick(this, event)"
                                class="px-1.5 py-0.5 text-xs rounded hover:brightness-110"
                                style="background-color: var(--bg-tertiary); color: var(--text-secondary);"
                                title="${this.t('folders.rename')}"
                            >✏️</button>
                            <button 
                                data-path="${esc(folder.path)}"
                                data-name="${esc(folder.name)}"
                                onclick="window.$root.handleDeleteFolderClick(this, event)"
                                class="px-1 py-0.5 text-xs rounded hover:brightness-110"
                                style="color: var(--error);"
                                title="${this.t('folders.delete')}"
                            >
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                                </svg>
                            </button>
                        </div>
                    </div>
            `;
            
            // If expanded, render folder contents (child folders + notes)
            if (isExpanded) {
                html += `<div class="folder-contents" style="padding-left: 10px;">`;
                
                // First, render child folders (if any)
                if (folder.children && Object.keys(folder.children).length > 0) {
                    const children = Object.entries(folder.children)
                        .filter(([k, v]) => !this.ui.hideUnderscoreFolders || !v.name.startsWith('_'))
                        .sort((a, b) => a[1].name.toLowerCase().localeCompare(b[1].name.toLowerCase()));
                    
                    children.forEach(([childKey, childFolder]) => {
                        html += this.renderFolderRecursive(childFolder, 0, false);
                    });
                }
                
                // Then, render notes and images in this folder (after subfolders)
                if (folder.notes && folder.notes.length > 0) {
                    const pagesShown = this.folderNotePages[folder.path] || 1;
                    const maxNotes = CONFIG.NOTES_PER_FOLDER_PAGE * pagesShown;
                    const totalNotes = folder.notes.length;
                    const visibleNotes = folder.notes.slice(0, maxNotes);
                    visibleNotes.forEach(note => {
                        html += this.renderNoteItem(note);
                    });
                    if (totalNotes > maxNotes) {
                        const remaining = totalNotes - maxNotes;
                        html += `<div class="px-2 py-1 text-xs" style="color: var(--text-tertiary); cursor: pointer;" onclick="window.$root.loadMoreFolderNotes('${esc(folder.path)}')">${this.t('notes.show_more', {count: remaining})}</div>`;
                    }
                }
                
                html += `</div>`; // Close folder-contents
            }
            
            html += `</div>`; // Close folder wrapper
            return html;
        },
        
        // Decode URL-encoded string (handles double-encoding as well)
        decodeNoteName(name) {
            if (!name) return '';
            try {
                // First decode
                let decoded = decodeURIComponent(name);
                // Handle double-encoding by decoding again if it still looks encoded
                if (decoded.includes('%')) {
                    try {
                        decoded = decodeURIComponent(decoded);
                    } catch (e) {
                        // If second decode fails, return the first decode result
                    }
                }
                return decoded;
            } catch (e) {
                // If decode fails, return original name
                return name;
            }
        },

        // Render a single note/media item (used by both folders and root level)
        renderNoteItem(note) {
            const esc = (s) => this.escapeHtmlAttr(s);
            const isMediaFile = note.type !== 'note';
            const isCurrentNote = this.note.current === note.path;
            const isCurrentMedia = this.media.current === note.path;
            const isCurrent = isMediaFile ? isCurrentMedia : isCurrentNote;

            // Decode note name to handle URL-encoded names (e.g., from double-encoding issues)
            const decodedName = this.decodeNoteName(note.name);

            // Share icon for shared notes
            const isShared = !isMediaFile && this.isNoteShared(note.path);
                            const shareIcon = isShared ? '<svg title="' + this.t('share.shared_label') + '" style="display: inline-block; width: 12px; height: 12px; vertical-align: middle; margin-right: 2px; opacity: 0.7;" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"></path></svg>' : '';            const icon = this.getMediaIcon(note.type);

            return `
                <div
                    data-path="${esc(note.path)}"
                    data-name="${esc(decodedName)}"
                    data-type="${note.type}"
                    draggable="true"
                    ondragstart="window.$root.onItemDragStart(this.dataset.path, this.dataset.type || 'note', event)"
                    ondragend="window.$root.onItemDragEnd()"
                    onclick="window.$root.handleItemClick(this)"
                    class="note-item px-2 py-1 text-sm relative"
                    style="${isCurrent ? 'background-color: var(--accent-light); color: var(--accent-primary);' : 'color: var(--text-primary);'} ${isMediaFile ? 'opacity: 0.85;' : ''} cursor: pointer;"
                    onmouseover="window.$root.handleItemHover(this, true)"
                    onmouseout="window.$root.handleItemHover(this, false)"
                >
                    <span class="truncate" style="display: block; padding-right: 30px;" title="${esc(decodedName)}">${shareIcon}${icon}${icon ? ' ' : ''}${esc(decodedName)}</span>
                    <button
                        data-path="${esc(note.path)}"
                        data-name="${esc(decodedName)}"
                        data-type="${note.type}"
                        onclick="window.$root.handleDeleteItemClick(this, event)"
                        class="note-delete-btn absolute right-2 top-1/2 transform -translate-y-1/2 px-1 py-0.5 text-xs rounded hover:brightness-110 transition-opacity"
                        style="opacity: 0; color: var(--error);"
                        title="${isMediaFile ? this.t('sidebar.delete_file') : this.t('sidebar.delete_note')}"
                    >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                        </svg>
                    </button>
                </div>
            `;
        },
        
        // Render root-level items (notes and media not in any folder)
        renderRootItems() {
            const root = this.folders.tree['__root__'];
            if (!root || !root.notes || root.notes.length === 0) {
                return '';
            }
            return root.notes.map(note => this.renderNoteItem(note)).join('');
        },
        
        // Toggle folder expansion
        toggleFolder(folderPath) {
            if (this.folders.expanded.has(folderPath)) {
                this.folders.expanded.delete(folderPath);
            } else {
                this.folders.expanded.add(folderPath);
            }
            // Force Alpine reactivity by creating new Set reference
            this.folders.expanded = new Set(this.folders.expanded);
        },
        
        // Check if folder is expanded
        isFolderExpanded(folderPath) {
            return this.folders.expanded.has(folderPath);
        },
        
        // Expand all folders
        expandAllFolders() {
            this.folders.all.forEach(folder => {
                this.folders.expanded.add(folder);
            });
            // Force Alpine reactivity
            this.folders.expanded = new Set(this.folders.expanded);
        },
        
        // Collapse all folders
        collapseAllFolders() {
            this.folders.expanded.clear();
            // Force Alpine reactivity
            this.folders.expanded = new Set(this.folders.expanded);
        },
        
        // Expand folder tree to show a specific note
        expandFolderForNote(notePath) {
            const parts = notePath.split('/');
            
            // If note is in root, no folders to expand
            if (parts.length <= 1) return;
            
            // Remove the note name (last part)
            parts.pop();
            
            // Build and expand all parent folders
            let currentPath = '';
            parts.forEach((part, index) => {
                currentPath = index === 0 ? part : `${currentPath}/${part}`;
                this.folders.expanded.add(currentPath);
            });
            
            // Force Alpine reactivity
            this.folders.expanded = new Set(this.folders.expanded);
        },
        
        // Scroll note into view in the sidebar navigation
        scrollNoteIntoView(notePath) {
            // Find the note element in the sidebar
            // Use a slight delay to ensure DOM is fully rendered with Alpine bindings applied
            setTimeout(() => {
                const sidebar = NoteCache.domCache.sidebar || document.querySelector('.flex-1.overflow-y-auto.custom-scrollbar');
                if (!sidebar) return;
                
                const noteElements = sidebar.querySelectorAll('.note-item');
                let targetElement = null;
                const noteName = notePath.split('/').pop().replace('.md', '');
                
                // Find the element that corresponds to this note
                noteElements.forEach(el => {
                    // Check if this is a note element (not folder) by checking if it has the note name
                    if (el.textContent.trim().startsWith(noteName) || el.textContent.includes(noteName)) {
                        // Check computed style to see if it's highlighted
                        const computedStyle = window.getComputedStyle(el);
                        const bgColor = computedStyle.backgroundColor;
                        
                        // Check if background has the accent color (not transparent or default)
                        if (bgColor && bgColor !== 'rgba(0, 0, 0, 0)' && bgColor !== 'transparent' && !bgColor.includes('255, 255, 255')) {
                            targetElement = el;
                        }
                    }
                });
                
                // If found, scroll it into view
                if (targetElement) {
                    targetElement.scrollIntoView({ 
                        behavior: 'smooth', 
                        block: 'center',
                        inline: 'nearest'
                    });
                }
            }, 200); // Increased delay to ensure Alpine has finished rendering
        },
        
        // Unified drag and drop handlers for notes, folders, and media
        onItemDragStart(itemPath, itemType, event) {
            // Set unified drag state
            this.drag.item = { path: itemPath, type: itemType };
            
            // Make drag image semi-transparent
            if (event.target) {
                event.target.style.opacity = '0.5';
            }
            
            event.dataTransfer.effectAllowed = 'all';
        },
        
        onItemDragEnd() {
            this.drag.item = null;
            this.drag.target = null;
            this.folders.dragOver = null;
            // Reset opacity of all draggable items
            document.querySelectorAll('.note-item, .folder-header').forEach(el => el.style.opacity = '1');
            // Reset drag-over class
            document.querySelectorAll('.drag-over').forEach(el => el.classList.remove('drag-over'));
        },
        
        
        // Handle dragover on editor to show cursor position
        onEditorDragOver(event) {
            if (!this.drag.item) return;
            
            event.preventDefault();
            this.drag.target = 'editor';
            
            // Focus the textarea
            const textarea = event.target;
            if (textarea.tagName !== 'TEXTAREA') return;
            
            textarea.focus();
            
            // Calculate cursor position from mouse coordinates
            const pos = this.getTextareaCursorFromPoint(textarea, event.clientX, event.clientY);
            if (pos >= 0) {
                textarea.setSelectionRange(pos, pos);
            }
        },
        
        // Calculate textarea cursor position from mouse coordinates
        getTextareaCursorFromPoint(textarea, x, y) {
            const rect = textarea.getBoundingClientRect();
            const style = window.getComputedStyle(textarea);
            const lineHeight = parseFloat(style.lineHeight) || parseFloat(style.fontSize) * 1.2;
            const paddingTop = parseFloat(style.paddingTop) || 0;
            const paddingLeft = parseFloat(style.paddingLeft) || 0;
            
            // Calculate which line we're on
            const relativeY = y - rect.top - paddingTop + textarea.scrollTop;
            const lineIndex = Math.max(0, Math.floor(relativeY / lineHeight));
            
            // Split content into lines
            const lines = textarea.value.split('\n');
            
            // Find the character position at the start of this line
            let charPos = 0;
            for (let i = 0; i < Math.min(lineIndex, lines.length); i++) {
                charPos += lines[i].length + 1; // +1 for newline
            }
            
            // If we're beyond the last line, position at end
            if (lineIndex >= lines.length) {
                return textarea.value.length;
            }
            
            // Approximate character position within the line based on X coordinate
            const relativeX = x - rect.left - paddingLeft;
            const charWidth = parseFloat(style.fontSize) * 0.6; // Approximate for monospace
            const charInLine = Math.max(0, Math.floor(relativeX / charWidth));
            const lineLength = lines[lineIndex]?.length || 0;
            
            return charPos + Math.min(charInLine, lineLength);
        },
        
        // Handle dragenter on editor
        onEditorDragEnter(event) {
            if (!this.drag.item) return;
            event.preventDefault();
            this.drag.target = 'editor';
        },
        
        // Handle dragleave on editor
        onEditorDragLeave(event) {
            // Only clear dropTarget if we're actually leaving the editor
            // (not just moving between child elements)
            if (event.target.tagName === 'TEXTAREA') {
                this.drag.target = null;
            }
        },
        
        // Handle drop into editor to create internal link or upload media
        async onEditorDrop(event) {
            event.preventDefault();
            this.drag.target = null;
            
            // Check if files are being dropped (media from file system)
            if (event.dataTransfer && event.dataTransfer.files && event.dataTransfer.files.length > 0) {
                await this.handleMediaDrop(event);
                return;
            }
            
            // Otherwise, handle note/media link drop from sidebar
            if (!this.drag.item) return;
            
            const notePath = this.drag.item.path;
            const isMediaFile = this.drag.item.type !== 'note';
            
            let link;
            if (isMediaFile) {
                // For media files (images, audio, video, PDF), use wiki-style embed link
                const filename = notePath.split('/').pop();
                link = `![[${filename}]]`;
            } else {
                // For notes, insert note link
                const noteName = notePath.split('/').pop().replace('.md', '');
                const encodedPath = notePath.split('/').map(segment => encodeURIComponent(segment)).join('/');
                link = `[${noteName}](${encodedPath})`;
            }
            
            // Insert at drop position
            const textarea = event.target;
            // Recalculate position from drop coordinates for accuracy
            let cursorPos = this.getTextareaCursorFromPoint(textarea, event.clientX, event.clientY);
            if (cursorPos < 0) cursorPos = textarea.selectionStart || 0;
            const textBefore = this.note.content.substring(0, cursorPos);
            const textAfter = this.note.content.substring(cursorPos);
            
            this.note.content = textBefore + link + textAfter;
            
            // Move cursor after the link
            this.$nextTick(() => {
                textarea.selectionStart = textarea.selectionEnd = cursorPos + link.length;
                textarea.focus();
            });
            
            // Trigger autosave
            this.autoSave();
            
            this.drag.item = null;
        },
        
        // Handle media files dropped into editor
        async handleMediaDrop(event) {
            if (!this.note.current) {
                this.showAlert(this.t('notes.open_first'));
                return;
            }
            
            const files = Array.from(event.dataTransfer.files);
            
            // Filter for allowed media types
            const allowedTypes = [
                // Images
                'image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp',
                // Audio
                'audio/mpeg', 'audio/mp3', 'audio/wav', 'audio/ogg', 'audio/m4a', 'audio/x-m4a',
                // Video
                'video/mp4', 'video/webm', 'video/quicktime',
                // Documents
                'application/pdf'
            ];
            const mediaFiles = files.filter(file => allowedTypes.includes(file.type.toLowerCase()));
            
            if (mediaFiles.length === 0) {
                this.showAlert(this.t('media.no_valid_files'));
                return;
            }
            
            const textarea = event.target;
            // Calculate cursor position from drop coordinates
            let cursorPos = this.getTextareaCursorFromPoint(textarea, event.clientX, event.clientY);
            if (cursorPos < 0) cursorPos = textarea.selectionStart || 0;
            
            // Upload each media file
            for (const file of mediaFiles) {
                try {
                    const mediaPath = await this.uploadMedia(file, this.note.current);
                    if (mediaPath) {
                        await this.insertMediaMarkdown(mediaPath, file.name, cursorPos);
                    }
                } catch (error) {
                    ErrorHandler.handle(`upload file ${file.name}`, error);
                }
            }
        },
        
        // Upload a media file (image, audio, video, PDF)
        async uploadMedia(file, notePath) {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('note_path', notePath);
            
            try {
                const response = await secureFetch('/api/upload-media', {
                    method: 'POST',
                    body: formData
                });
                
                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(error.detail || 'Upload failed');
                }
                
                const data = await response.json();
                return data.path;
            } catch (error) {
                throw error;
            }
        },
        
        // Insert media markdown at cursor position using wiki-style syntax
        // This ensures media links don't break when notes are moved
        async insertMediaMarkdown(mediaPath, altText, cursorPos) {
            // Extract just the filename from the path (e.g., "folder/_attachments/image.png" -> "image.png")
            const filename = mediaPath.split('/').pop();
            
            // Use wiki-style embed link: ![[filename.png]] or ![[filename.png|alt text]]
            // The alt text is optional - only add if different from filename
            const filenameWithoutExt = filename.replace(/\.[^/.]+$/, '');
            const altWithoutExt = altText.replace(/\.[^/.]+$/, '');
            
            // If alt text is meaningful (not just "pasted-image"), include it
            const markdown = (altWithoutExt && altWithoutExt !== filenameWithoutExt && !altWithoutExt.startsWith('pasted-image'))
                ? `![[${filename}|${altWithoutExt}]]`
                : `![[${filename}]]`;
            
            // Add the newly uploaded media to the lookup map immediately
            // This ensures the preview can render it without waiting for loadNotes()
            const filenameLower = filename.toLowerCase();
            if (!NoteCache.mediaLookup.has(filenameLower)) {
                NoteCache.mediaLookup.set(filenameLower, mediaPath);
            }
            
            // Reload notes to sync with backend (may be delayed due to backend cache)
            await this.loadNotes();
            
            const textBefore = this.note.content.substring(0, cursorPos);
            const textAfter = this.note.content.substring(cursorPos);
            
            this.note.content = textBefore + markdown + '\n' + textAfter;
            
            // Trigger autosave
            this.autoSave();
            
            // Invalidate preview cache to force immediate update
            NoteCache.lastRenderedContent = undefined;
            NoteCache.cachedRenderedHTML = undefined;
        },
        
        // Handle paste event for clipboard media (images) and URL-to-Markdown conversion
        async handlePaste(event) {
            if (!this.note.current) return;

            // First, try to convert rich-text URLs (with titles) to Markdown links
            // This handles clipboard from browsers like Edge/Chrome that copy links as HTML
            const htmlContent = event.clipboardData?.getData('text/html');
            if (htmlContent) {
                const converted = this.convertHtmlLinksToMarkdown(htmlContent);
                if (converted) {
                    event.preventDefault();
                    const textarea = event.target;
                    const cursorPos = textarea.selectionStart || 0;
                    const textBefore = this.note.content.substring(0, cursorPos);
                    const textAfter = this.note.content.substring(cursorPos);
                    this.note.content = textBefore + converted + textAfter;
                    // Trigger autosave and invalidate preview cache
                    this.autoSave();
                    NoteCache.lastRenderedContent = undefined;
                    NoteCache.cachedRenderedHTML = undefined;
                    // Set cursor position after inserted text
                    requestAnimationFrame(() => {
                        const newPos = cursorPos + converted.length;
                        textarea.setSelectionRange(newPos, newPos);
                    });
                    return;
                }
            }

            const items = event.clipboardData?.items;
            if (!items) return;

            for (const item of items) {
                if (item.type.startsWith('image/')) {
                    event.preventDefault();

                    const blob = item.getAsFile();
                    if (blob) {
                        try {
                            const textarea = event.target;
                            const cursorPos = textarea.selectionStart || 0;

                            // Create a simple filename - backend will add timestamp to prevent collisions
                            const ext = item.type.split('/')[1] || 'png';
                            const filename = `pasted-image.${ext}`;

                            // Create a File from the blob
                            const file = new File([blob], filename, { type: item.type });

                            const mediaPath = await this.uploadMedia(file, this.note.current);
                            if (mediaPath) {
                                await this.insertMediaMarkdown(mediaPath, filename, cursorPos);
                            }
                        } catch (error) {
                            ErrorHandler.handle('paste media', error);
                        }
                    }
                    break; // Only handle first media item
                }
            }
        },

        // Convert HTML anchor tags from clipboard to Markdown links
        // e.g., <a href="https://example.com">Page Title</a> -> [Page Title](https://example.com)
        convertHtmlLinksToMarkdown(html) {
            // Parse the HTML content
            const parser = new DOMParser();
            const doc = parser.parseFromString(html, 'text/html');

            // Check if the clipboard contains anchor tags with href
            const anchors = doc.body.querySelectorAll('a[href]');
            if (anchors.length === 0) {
                return null;
            }

            // Convert the HTML body content, replacing anchor tags with Markdown links
            let bodyHtml = doc.body.innerHTML;

            // Replace all anchor tags with Markdown format
            bodyHtml = bodyHtml.replace(/<a\s+[^>]*href=["']([^"']*)["'][^>]*>(.*?)<\/a>/gi, (match, href, innerHtml) => {
                // Skip anchors that contain other block elements or complex nested structures
                if (/<(div|p|table|ul|ol|img|video|audio)/i.test(innerHtml)) {
                    return match; // Leave complex anchors as plain text
                }

                // Strip any remaining HTML from the inner text
                const tempDiv = document.createElement('div');
                tempDiv.innerHTML = innerHtml;
                const title = tempDiv.textContent.trim() || href;

                return `[${title}](${href})`;
            });

            // Extract just the text content from the converted HTML
            const tempBody = document.createElement('div');
            tempBody.innerHTML = bodyHtml;
            let textContent = tempBody.textContent;

            // Clean up excessive whitespace but preserve structure
            textContent = textContent
                .replace(/\r\n/g, '\n')
                .replace(/\n{3,}/g, '\n\n')  // Max 2 consecutive newlines
                .trim();

            // If the result is empty or just a single simple URL, return null to fall through to default paste
            if (!textContent) {
                return null;
            }

            // Check if it's just a bare URL (no title found) - let default paste handle it
            const markdownLinkRegex = /\[.+\]\(.+\)/;
            if (!markdownLinkRegex.test(textContent) && /^https?:\/\//.test(textContent)) {
                return null; // Bare URL, let browser default handle it
            }

            return textContent;
        },
        
        // Media type detection based on file extension
        getMediaType(filename) {
            if (!filename) return null;
            const ext = filename.split('.').pop().toLowerCase();
            const mediaTypes = {
                image: ['jpg', 'jpeg', 'png', 'gif', 'webp'],
                audio: ['mp3', 'wav', 'ogg', 'm4a'],
                video: ['mp4', 'webm', 'mov', 'avi'],
                document: ['pdf'],
            };
            for (const [type, extensions] of Object.entries(mediaTypes)) {
                if (extensions.includes(ext)) return type;
            }
            return null;
        },
        
        // Get icon for media type
        getMediaIcon(type) {
            const icons = {
                image: '🖼️',
                audio: '🎵',
                video: '🎬',
                document: '📄',
            };
            return icons[type] || '';
        },
        
        // Open a note or media file (unified handler for sidebar/homepage clicks)
        openItem(path, type = 'note', searchHighlight = '', lineNumber = 0) {
            this.graph.show = false;
            
            // Check if it's a media file by type or extension
            const mediaType = type !== 'note' ? type : this.getMediaType(path);
            if (mediaType && mediaType !== 'note') {
                this.viewMedia(path, mediaType);
                return;
            }
            
            // If clicking the same note that's already open, just highlight and scroll
            if (this.note.current === path) {
                if (searchHighlight) {
                    this.search.highlight = searchHighlight;
                    this.search.targetLineNumber = lineNumber;
                    this.$nextTick(() => {
                        this.highlightSearchTerm(searchHighlight, true);
                    });
                }
                return;
            }
            
            // Load different note
            this.loadNote(path, true, searchHighlight, lineNumber);
        },
        
        // View a media file (image, audio, video, PDF) in the main pane
        viewMedia(mediaPath, mediaType = null, updateHistory = true) {
            this.graph.show = false; // Ensure graph is closed
            this.note.current = '';
            this.note.name = '';
            this.note.content = '';
            this.media.current = mediaPath; // Reuse currentMedia for all media
            this.media.type = mediaType || this.getMediaType(mediaPath) || 'image';
            this.modals.share.info = null; // Reset share info
            this.ui.viewMode = 'preview'; // Use preview mode to show media
            
            // Update browser tab title
            const fileName = mediaPath.split('/').pop();
            document.title = `${fileName} - ${this.app.name}`;
            
            // Expand folder tree to show the media file
            this.expandFolderForNote(mediaPath);
            
            // Update browser URL
            if (updateHistory) {
                // Encode each path segment to handle special characters
                const encodedPath = mediaPath.split('/').map(segment => encodeURIComponent(segment)).join('/');
                window.history.pushState(
                    { mediaPath: mediaPath },
                    '',
                    `/${encodedPath}`
                );
            }
        },
        
        // Backward compatibility alias
        viewImage(mediaPath, updateHistory = true) {
            this.viewMedia(mediaPath, 'image', updateHistory);
        },
        
        // Delete a media file (image, audio, video, PDF)
        async deleteMedia(mediaPath) {
            const filename = mediaPath.split('/').pop();
            if (!await this.showConfirm(this.t('media.confirm_delete', { name: filename }))) return;
            
            try {
                const response = await secureFetch(`/api/notes/${encodeURIComponent(mediaPath)}`, {
                    method: 'DELETE'
                });
                
                if (response.ok) {
                    await this.loadNotes(); // Refresh tree
                    
                    // Clear viewer if deleting currently viewed media
                    if (this.media.current === mediaPath) {
                        this.media.current = '';
                    }
                } else {
                    throw new Error('Failed to delete media file');
                }
            } catch (error) {
                ErrorHandler.handle('delete media', error);
            }
        },
        
        // Handle clicks on internal links and checkboxes in preview
        async handleInternalLink(event) {
            // Check if clicked element is an interactive checkbox
            const checkbox = event.target.closest('input[data-interactive-checkbox]');
            if (checkbox) {
                // Don't preventDefault - let the browser toggle the checkbox naturally.
                // We read the NEW state after the browser has toggled it.
                event.stopPropagation();
                this.toggleTaskCheckbox(checkbox);
                return;
            }

            // Check if clicked element is a link
            const link = event.target.closest('a');
            if (!link) return;
            
            const href = link.getAttribute('href');
            if (!href) return;
            
            // Check if it's an external link or API path (media files, etc.)
            if (href.startsWith('http://') || href.startsWith('https://') || href.startsWith('//') || href.startsWith('mailto:') || href.startsWith('/api/')) {
                return; // Let external links and API paths work normally
            }
            
            // Prevent default navigation for internal links
            event.preventDefault();
            
            // Parse href into note path and anchor (e.g., "note.md#section" -> notePath="note.md", anchor="section")
            const decodedHref = decodeURIComponent(href);
            const hashIndex = decodedHref.indexOf('#');
            const notePath = hashIndex !== -1 ? decodedHref.substring(0, hashIndex) : decodedHref;
            const anchor = hashIndex !== -1 ? decodedHref.substring(hashIndex + 1) : null;
            
            // If it's just an anchor link (#heading), scroll within current note
            if (!notePath && anchor) {
                this.scrollToAnchor(anchor);
                return;
            }
            
            // Skip if no path
            if (!notePath) return;
            
            // Find the note by path (try exact match first, then with .md extension)
            let targetNote = this.notes.find(n => 
                n.path === notePath || 
                n.path === notePath + '.md'
            );
            
            if (!targetNote) {
                // Try to find by name (in case link uses just the note name without path)
                targetNote = this.notes.find(n => 
                    n.name === notePath || 
                    n.name === notePath + '.md' ||
                    n.name.toLowerCase() === notePath.toLowerCase() ||
                    n.name.toLowerCase() === (notePath + '.md').toLowerCase()
                );
            }
            
            if (!targetNote) {
                // Last resort: case-insensitive path matching
                targetNote = this.notes.find(n => 
                    n.path.toLowerCase() === notePath.toLowerCase() ||
                    n.path.toLowerCase() === (notePath + '.md').toLowerCase()
                );
            }
            
            if (targetNote) {
                // Load the note, then scroll to anchor if present
                this.loadNote(targetNote.path).then(() => {
                    if (anchor) {
                        // Small delay to ensure content is rendered
                        setTimeout(() => this.scrollToAnchor(anchor), 100);
                    }
                });
            } else if (await this.showConfirm(this.t('notes.create_from_link', { path: notePath }))) {
                // Note doesn't exist - create it (reuses createNote with duplicate check)
                this.createNote(null, notePath);
            }
        },
        
        // Scroll to an anchor (heading) by slug - reuses outline data
        scrollToAnchor(anchor) {
            // Normalize the anchor (GitHub-style slug)
            const targetSlug = anchor
                .toLowerCase()
                .replace(/[^\w\s-]/g, '')
                .replace(/\s+/g, '-')
                .replace(/-+/g, '-');
            
            // Find matching heading in outline
            const heading = this.note.outline.find(h => h.slug === targetSlug);
            
            if (heading) {
                this.scrollToHeading(heading);
            } else {
                // Fallback: try to find heading by exact text match
                const headingByText = this.note.outline.find(h => 
                    h.text.toLowerCase().replace(/\s+/g, '-') === anchor.toLowerCase()
                );
                if (headingByText) {
                    this.scrollToHeading(headingByText);
                }
            }
        },

        // Toggle a task checkbox in the note content
        toggleTaskCheckbox(checkboxEl) {
            if (!this.note.current || !this.note.content) return;

            // The browser has already toggled the checkbox state.
            // checkboxEl.checked is the new (toggled) state.
            const newChecked = checkboxEl.checked;

            // Find the ordinal position of this checkbox among all task checkboxes
            // in the rendered preview
            const preview = checkboxEl.closest('.markdown-preview');
            if (!preview) return;

            const allCheckboxes = preview.querySelectorAll('input[data-interactive-checkbox]');
            let checkboxIndex = -1;
            allCheckboxes.forEach((cb, idx) => {
                if (cb === checkboxEl) checkboxIndex = idx;
            });

            if (checkboxIndex === -1) return;

            // Find the corresponding task line in note.content by matching ordinal position.
            // IMPORTANT: Skip lines inside fenced code blocks (``` ... ```) and
            // indented code blocks (4+ spaces) because they are NOT rendered as
            // checkboxes by marked (replaced with placeholders or <pre><code>).
            const lines = this.note.content.split('\n');
            // Support both unordered (-, *, +) and ordered (1., 2., etc.) list markers
            // Also match uppercase [X] (GFM spec allows both [x] and [X])
            const taskLineRegex = /^(\s*(?:[*\-+]|\d+\.)\s*)\[([ xX])\]/;
            let taskMatches = [];
            let inFencedCodeBlock = false;

            for (let i = 0; i < lines.length; i++) {
                const line = lines[i];

                // Track fenced code block boundaries (GFM allows 0-3 leading spaces before ```)
                if (/^\s{0,3}(`{3,}|~{3,})/.test(line)) {
                    inFencedCodeBlock = !inFencedCodeBlock;
                    continue;
                }

                // Skip lines inside fenced code blocks
                if (inFencedCodeBlock) continue;

                // Skip indented code block lines (4+ spaces or 1+ tab at start)
                // but NOT nested list items like "    - [ ]" which marked renders as checkboxes
                // Only skip if it looks like code: indented but NOT starting with a list marker
                if (/^(    |\t)+/.test(line) && !/^(    |\t)+\s*[*\-+]/.test(line) && !/^(    |\t)+\d+\.\s/.test(line) && line.trim() !== '') {
                    continue;
                }

                // Check if this line is a task list item
                const match = line.match(taskLineRegex);
                if (match) {
                    taskMatches.push({
                        lineIndex: i,
                        prefix: match[1],       // e.g., "- " or "1. "
                        checked: match[2],      // " " or "x" or "X"
                        fullMatch: match[0],    // e.g., "- [ ]"
                        fullLine: line,
                    });
                }
            }

            if (checkboxIndex >= taskMatches.length) return;

            const targetMatch = taskMatches[checkboxIndex];

            // Build the replacement with the new checked state
            const checkedChar = newChecked ? 'x' : ' ';
            const newTaskPart = targetMatch.fullMatch.replace(
                /\[[ xX]\]/,
                `[${checkedChar}]`
            );

            // Update the specific line in content
            const newLines = this.note.content.split('\n');
            newLines[targetMatch.lineIndex] = targetMatch.fullLine.replace(
                targetMatch.fullMatch,
                newTaskPart
            );
            this.note.content = newLines.join('\n');

            // Trigger auto-save
            this.autoSave();
        },


        cancelDrag() {
            // Cancel any active drag operation (triggered by ESC key)
            this.drag.item = null;
            this.drag.target = null;
            this.folders.dragOver = null;
            // Reset styles - only query elements with drag-over class (more efficient)
            document.querySelectorAll('.folder-item').forEach(el => el.style.opacity = '1');
            document.querySelectorAll('.note-item').forEach(el => el.style.opacity = '1');
            document.querySelectorAll('.drag-over').forEach(el => el.classList.remove('drag-over'));
        },
        
        async onFolderDrop(targetFolderPath) {
            // Ignore if we're dropping into the editor
            if (this.drag.target === 'editor') {
                return;
            }
            
            // Capture dragged item info immediately (ondragend may clear it)
            if (!this.drag.item) return;
            const { path: draggedPath, type: draggedType } = this.drag.item;
            
            // Determine item category for endpoint selection
            const isFolder = draggedType === 'folder';
            const isNote = draggedType === 'note';
            const isMedia = !isFolder && !isNote; // image, audio, video, document
            
            // Handle folder drop
            if (isFolder) {
                // Decode draggedPath to handle any URL-encoded characters
                // This ensures we work with the actual filesystem path, not encoded version
                const decodedDraggedPath = this.decodeNoteName(draggedPath);
                const folderName = decodedDraggedPath.split('/').pop();
                
                // Prevent dropping folder into itself or its subfolders
                if (targetFolderPath === decodedDraggedPath || 
                    targetFolderPath.startsWith(decodedDraggedPath + '/')) {
                    this.showAlert(this.t('folders.cannot_move_into_self'));
                    return;
                }
                
                const newPath = targetFolderPath ? `${targetFolderPath}/${folderName}` : folderName;
                
                if (newPath === decodedDraggedPath) return;
                
                // Capture favorites info before async call
                const oldPrefix = decodedDraggedPath + '/';
                const newPrefix = newPath + '/';
                
                try {
                    const response = await secureFetch('/api/folders/move', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ oldPath: decodedDraggedPath, newPath })
                    });
                    
                    if (response.ok) {
                        // Update favorites for notes inside moved folder
                        const favoritesInFolder = this._favoritesState.list.filter(f => f.startsWith(oldPrefix));
                        if (favoritesInFolder.length > 0) {
                            const newFavorites = this._favoritesState.list.map(f => 
                                f.startsWith(oldPrefix) ? newPrefix + f.substring(oldPrefix.length) : f
                            );
                            this._favoritesState.list = newFavorites;
                            this._favoritesState.set = new Set(newFavorites);
                            this.saveFavorites();
                        }
                        
                        // Keep folder expanded if it was
                        const wasExpanded = this.folders.expanded.has(decodedDraggedPath);
                        
                        await this.loadNotes();
                        await this.loadSharedNotePaths();
                        
                        if (wasExpanded) {
                            this.folders.expanded.delete(decodedDraggedPath);
                            this.folders.expanded.add(newPath);
                            this.saveExpandedFolders();
                        }
                    } else {
                        const errorData = await response.json().catch(() => ({}));
                        this.showAlert(errorData.detail || this.t('move.failed_folder'));
                    }
                } catch (error) {
                    console.error('Failed to move folder:', error);
                    this.showAlert(this.t('move.failed_folder'));
                }
                return;
            }
            
            // Handle note or media drop into folder
            const decodedDraggedPath = this.decodeNoteName(draggedPath);
            const item = this.notes.find(n => n.path === decodedDraggedPath);
            if (!item) return;

            // Decode draggedPath to handle any URL-encoded characters
            // This ensures we work with the actual filesystem path, not encoded version
            const filename = decodedDraggedPath.split('/').pop();
            const newPath = targetFolderPath ? `${targetFolderPath}/${filename}` : filename;
            
            if (newPath === decodedDraggedPath) return;
            
            // Check if note is favorited (only for notes)
            const wasFavorited = isNote && this._favoritesState.set.has(decodedDraggedPath);
            
            try {
                // Use different endpoint for media vs notes
                const endpoint = isMedia ? '/api/media/move' : '/api/notes/move';
                const response = await secureFetch(endpoint, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ oldPath: decodedDraggedPath, newPath })
                });
                
                if (response.ok) {
                    // Clear draft for old path
                    if (isNote) this.clearDraft(decodedDraggedPath);
                    
                    // Update favorites if the moved note was favorited
                    if (wasFavorited) {
                        const newFavorites = this._favoritesState.list.map(f => f === decodedDraggedPath ? newPath : f);
                        this._favoritesState.list = newFavorites;
                        this._favoritesState.set = new Set(newFavorites);
                        this.saveFavorites();
                    }
                    
                    // Keep current item open if it was the moved one
                    const wasCurrentNote = this.note.current === decodedDraggedPath;
                    const wasCurrentMedia = this.media.current === decodedDraggedPath;
                    
                    await this.loadNotes();
                    if (isNote) {
                        await this.loadSharedNotePaths();
                    }
                    
                    if (wasCurrentNote) this.note.current = newPath;
                    if (wasCurrentMedia) this.media.current = newPath;
                } else {
                    const errorData = await response.json().catch(() => ({}));
                    const errorKey = isMedia ? 'move.failed_media' : 'move.failed_note';
                    this.showAlert(errorData.detail || this.t(errorKey));
                    return;
                }
            } catch (error) {
                console.error(`Failed to move ${isMedia ? 'media' : 'note'}:`, error);
                const errorKey = isMedia ? 'move.failed_media' : 'move.failed_note';
                this.showAlert(this.t(errorKey));
            }
        },
        
        
        // Load a specific note
        async loadNote(notePath, updateHistory = true, searchQuery = '', lineNumber = 0) {
            try {
                // Save scroll position of current note before switching
                if (this.note.current && this.note.current !== notePath) {
                    this.saveScrollPosition();
                }
                
                // Close mobile sidebar when a note is selected
                this.ui.mobileSidebarOpen = false;
                
                const encodedPath = notePath.split('/').map(segment => encodeURIComponent(segment)).join('/');
                const response = await fetch(`/api/notes/${encodedPath}`);
                
                // Check if note exists
                if (!response.ok) {
                    if (response.status === 404) {
                        // Note not found - silently redirect to home
                        window.history.replaceState({ homepageFolder: this.homepage.selectedFolder || '' }, '', '/');
                        this.note.current = '';
                        this.note.content = '';
                        this.media.current = '';
                        document.title = this.app.name;
                        return;
                    }
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                
                const data = await response.json();
                
                this.note.current = notePath;
                NoteCache.lastRenderedContent = ''; // Clear render cache for new note
                NoteCache.cachedRenderedHTML = '';
                NoteCache.initializedVideoSources = new Set(); // Clear video cache for new note
                this.note.content = data.content;
                this.note.name = this.decodeNoteName(notePath.split('/').pop().replace('.md', ''));
                this.media.current = ''; // Clear image viewer when loading a note
                this.modals.share.info = null; // Reset share info for new note
                this.note.backlinks = []; // Clear backlinks cache for new note
                this.note.modified = data.metadata ? data.metadata.modified : '';
                this.note.dirty = false;
                
                // Update browser tab title
                document.title = `${this.note.name} - ${this.app.name}`;
                this.note.lastSaved = false;
                
                // Extract outline for TOC panel
                this.extractOutline(data.content);
                
                // Initialize undo/redo history for this note (with cursor at start)
                this.history.undo = [{ content: data.content, cursorPos: 0 }];
                this.history.redo = [];
                this.history.hasPendingChanges = false;
                
                // Update browser URL and history
                if (updateHistory) {
                    // Encode the path properly (spaces become %20, etc.)
                    const pathWithoutExtension = notePath.replace('.md', '');
                    // Encode each path segment to handle special characters
                    const encodedPath = pathWithoutExtension.split('/').map(segment => encodeURIComponent(segment)).join('/');
                    let url = `/${encodedPath}`;
                    // Add search query parameter if present
                    if (searchQuery) {
                        url += `?search=${encodeURIComponent(searchQuery)}`;
                    }
                    window.history.pushState(
                        { 
                            notePath: notePath, 
                            searchQuery: searchQuery,
                            homepageFolder: this.homepage.selectedFolder || '' // Save current folder state
                        },
                        '',
                        url
                    );
                }
                
                // Calculate stats if enabled
                if (this.stats.enabled) {
                    this.calculateStats();
                }
                
                // Parse frontmatter metadata
                this.parseMetadata();
                
                this.loadAttachments();
                
                // Store search query and target line number for highlighting
                if (searchQuery) {
                    this.search.highlight = searchQuery;
                    this.search.targetLineNumber = lineNumber; // Store line number from search result
                } else {
                    // Clear highlights if no search query
                    this.search.highlight = '';
                    this.search.targetLineNumber = 0;
                }
                
                // Expand folder tree to show the loaded note
                this.expandFolderForNote(notePath);
                
                // Use $nextTick twice to ensure Alpine.js has time to:
                // 1. First tick: expand folders and update DOM
                // 2. Second tick: highlight the note and setup everything else
                this.$nextTick(() => {
                    this.$nextTick(() => {
                        this.refreshDOMCache();
                        this.setupScrollSync();
                        
                        // Only restore scroll position if NOT from search result
                        // Search results will scroll to the matched line via highlightSearchTerm
                        if (!searchQuery) {
                            this.restoreScrollPosition();
                        }
                        
                        // Apply or clear search highlighting
                        if (searchQuery) {
                            // Pass true to focus editor when loading from search result
                            this.highlightSearchTerm(searchQuery, true);
                        } else {
                            this.clearSearchHighlights();
                        }
                        
                        // Scroll note into view in sidebar if needed
                        this.scrollNoteIntoView(notePath);
                    });
                });
                
            } catch (error) {
                ErrorHandler.handle('load note', error);
            }
        },

        // Load backlinks for the current note
        async loadBacklinks() {
            if (!this.note.current) {
                this.note.backlinks = [];
                return;
            }

            this.note.backlinksLoading = true;

            try {
                const encodedPath = this.note.current.split('/').map(segment => encodeURIComponent(segment)).join('/');
                const response = await fetch(`/api/backlinks/${encodedPath}`);

                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }

                const data = await response.json();
                this.note.backlinks = data.backlinks || [];
            } catch (error) {
                console.error('Failed to load backlinks:', error);
                this.note.backlinks = [];
            } finally {
                this.note.backlinksLoading = false;
            }
        },

        // Scroll to a specific line in the editor
        scrollToLine(lineNumber) {
            const editor = NoteCache.domCache.editor || document.querySelector('.editor-textarea');
            if (!editor) return;

            const lines = this.note.content.split('\n');
            if (lineNumber < 1 || lineNumber > lines.length) return;

            // Calculate character position
            let charPos = 0;
            for (let i = 0; i < lineNumber - 1 && i < lines.length; i++) {
                charPos += lines[i].length + 1;
            }

            // Set cursor position
            editor.focus();
            editor.setSelectionRange(charPos, charPos);

            // Calculate scroll position
            const lineHeight = parseFloat(getComputedStyle(editor).lineHeight) || 24;
            const scrollTop = (lineNumber - 3) * lineHeight - editor.clientHeight / 3;
            editor.scrollTop = Math.max(0, scrollTop);
        },

        // Load item (note or media) from URL path
        loadItemFromURL() {
            // Get path from URL (e.g., /folder/note or /folder/image.png)
            let path = window.location.pathname;
            
            // Strip .md extension if present (for MKdocs/Zensical integration)
            if (path.toLowerCase().endsWith('.md')) {
                path = path.slice(0, -3);
                // Update URL bar to show clean path without .md
                window.history.replaceState(null, '', path);
            }
            
            // Skip if root path or static assets
            if (path === '/' || path.startsWith('/static/') || path.startsWith('/api/')) {
                return;
            }
            
            // Remove leading slash and decode URL encoding (e.g., %20 -> space)
            const decodedPath = decodeURIComponent(path.substring(1));
            
            // Check if this is a media file (image, audio, video, PDF)
            const matchedItem = this.notes.find(n => n.path === decodedPath);
            
            if (matchedItem && matchedItem.type !== 'note') {
                // It's a media file, view it
                this.viewMedia(decodedPath, matchedItem.type, false); // false = don't update history
            } else {
                // It's a note, add .md extension and load it
                const notePath = decodedPath + '.md';
                
                // Parse query string for search parameter
                const urlParams = new URLSearchParams(window.location.search);
                const searchParam = urlParams.get('search');
                
                // Try to load the note directly - the backend will handle 404 if it doesn't exist
                // This is more robust than checking the frontend notes list
                this.loadNote(notePath, false, searchParam || '');
                
                // If there's a search parameter, populate the search box and trigger search
                if (searchParam) {
                    this.search.query = searchParam;
                    // Trigger search to populate results list
                    this.searchNotes();
                }
            }
        },
        
        // Highlight search term in editor and preview
        highlightSearchTerm(query, focusEditor = false) {
            if (!query || !query.trim()) {
                this.clearSearchHighlights();
                return;
            }
            
            const searchTerm = query.trim();
            
            // Highlight in editor (textarea) - find match positions
            this.highlightInEditor(searchTerm, focusEditor);
            
            // Highlight in preview (rendered HTML)
            this.highlightInPreview(searchTerm);
        },
        
        // Highlight search term in the editor textarea
        highlightInEditor(searchTerm, shouldFocus = false) {
            const editor = NoteCache.domCache.editor || document.getElementById('note-editor');
            if (!editor) return;
            
            // Use note.content directly - more reliable than editor.value during load
            const content = this.note.content || '';
            const lowerContent = content.toLowerCase();
            const lowerTerm = searchTerm.toLowerCase();
            
            // Find all match positions in editor content
            this.search.editorMatchPositions = [];
            let searchPos = 0;
            let matchIndex;
            
            while ((matchIndex = lowerContent.indexOf(lowerTerm, searchPos)) !== -1) {
                this.search.editorMatchPositions.push({
                    start: matchIndex,
                    end: matchIndex + searchTerm.length,
                    text: content.substring(matchIndex, matchIndex + searchTerm.length)
                });
                searchPos = matchIndex + 1;
            }
            
            // Update total matches count (combine editor and preview matches)
            // Preview matches will be counted separately in highlightInPreview
            this.search.editorMatchCount = this.search.editorMatchPositions.length;
            
            // Set initial match index if we have matches
            if (this.search.editorMatchPositions.length > 0) {
                this.search.currentEditorMatchIndex = 0;
            } else {
                this.search.currentEditorMatchIndex = -1;
            }
            
            // Scroll to match if any
            if (this.search.editorMatchPositions.length > 0 && shouldFocus) {
                // If we have a target line number from search result, find the match on that line
                if (this.search.targetLineNumber && this.search.targetLineNumber > 0) {
                    const targetLine = this.search.targetLineNumber;
                    // Find the match position that corresponds to the target line
                    let targetMatchIndex = 0;
                    for (let i = 0; i < this.search.editorMatchPositions.length; i++) {
                        const match = this.search.editorMatchPositions[i];
                        const textBefore = content.substring(0, match.start);
                        const lineOfMatch = textBefore.split('\n').length;
                        if (lineOfMatch === targetLine) {
                            targetMatchIndex = i;
                            break;
                        }
                        // If we passed the target line, use the closest match
                        if (lineOfMatch > targetLine) {
                            targetMatchIndex = i > 0 ? i - 1 : 0;
                            break;
                        }
                    }
                    this.scrollToEditorMatch(targetMatchIndex);
                } else {
                    // No target line, scroll to first match
                    this.scrollToEditorMatch(0);
                }
            }
        },
        
        // Scroll to a specific match in the editor
        scrollToEditorMatch(index) {
            const editor = NoteCache.domCache.editor || document.getElementById('note-editor');
            if (!editor || !this.search.editorMatchPositions || index < 0 || index >= this.search.editorMatchPositions.length) return;
            
            const match = this.search.editorMatchPositions[index];
            
            // Use note.content for line calculation (more reliable during load)
            const content = this.note.content || '';
            const textBefore = content.substring(0, match.start);
            const lineNumber = textBefore.split('\n').length;
            
            // Get line height from computed style
            const computedStyle = window.getComputedStyle(editor);
            const lineHeight = parseFloat(computedStyle.lineHeight) || 20;
            const paddingTop = parseFloat(computedStyle.paddingTop) || 24;
            
            // Scroll to show the match with some context
            editor.scrollTop = (lineNumber - 3) * lineHeight - paddingTop;
            
            // Use requestAnimationFrame to ensure DOM is ready before setting selection
            requestAnimationFrame(() => {
                editor.focus();
                editor.setSelectionRange(match.start, match.end);
            });
            
            // Store current editor match index
            this.search.currentEditorMatchIndex = index;
        },
        
        // Navigate to next match in editor
        nextEditorMatch() {
            if (!this.search.editorMatchPositions || this.search.editorMatchPositions.length === 0) return;
            this.search.currentEditorMatchIndex = (this.search.currentEditorMatchIndex + 1) % this.search.editorMatchPositions.length;
            this.scrollToEditorMatch(this.search.currentEditorMatchIndex);
        },
        
        // Navigate to previous match in editor
        previousEditorMatch() {
            if (!this.search.editorMatchPositions || this.search.editorMatchPositions.length === 0) return;
            this.search.currentEditorMatchIndex = (this.search.currentEditorMatchIndex - 1 + this.search.editorMatchPositions.length) % this.search.editorMatchPositions.length;
            this.scrollToEditorMatch(this.search.currentEditorMatchIndex);
        },
        
        // Highlight search term in the preview pane
        highlightInPreview(searchTerm) {
            const preview = NoteCache.domCache.previewContent || document.querySelector('.markdown-preview');
            if (!preview) return;
            
            // Remove existing preview highlights (without resetting editor match state)
            this.clearPreviewHighlights();
            
            // Create a tree walker to find all text nodes
            const walker = document.createTreeWalker(
                preview,
                NodeFilter.SHOW_TEXT,
                null,
                false
            );
            
            const textNodes = [];
            let node;
            while (node = walker.nextNode()) {
                // Skip code blocks and pre tags
                if (node.parentElement.tagName === 'CODE' || 
                    node.parentElement.tagName === 'PRE') {
                    continue;
                }
                textNodes.push(node);
            }
            
            const lowerTerm = searchTerm.toLowerCase();
            let matchIndex = 0;
            
            // Highlight matches in text nodes
            textNodes.forEach(textNode => {
                const text = textNode.textContent;
                const lowerText = text.toLowerCase();
                
                if (lowerText.includes(lowerTerm)) {
                    const fragment = document.createDocumentFragment();
                    let lastIndex = 0;
                    let index;
                    
                    while ((index = lowerText.indexOf(lowerTerm, lastIndex)) !== -1) {
                        // Add text before match
                        if (index > lastIndex) {
                            fragment.appendChild(
                                document.createTextNode(text.substring(lastIndex, index))
                            );
                        }
                        
                        // Add highlighted match
                        const mark = document.createElement('mark');
                        mark.className = 'search-highlight';
                        mark.setAttribute('data-match-index', matchIndex);
                        mark.textContent = text.substring(index, index + searchTerm.length);
                        
                        // First match is active (styled via CSS)
                        if (matchIndex === 0) {
                            mark.classList.add('active-match');
                        }
                        
                        fragment.appendChild(mark);
                        matchIndex++;
                        
                        lastIndex = index + searchTerm.length;
                    }
                    
                    // Add remaining text
                    if (lastIndex < text.length) {
                        fragment.appendChild(
                            document.createTextNode(text.substring(lastIndex))
                        );
                    }
                    
                    // Replace text node with highlighted fragment
                    textNode.parentNode.replaceChild(fragment, textNode);
                }
            });
            
            // Update total matches and reset current index
            this.search.totalMatches = matchIndex;
            this.search.matchIndex = matchIndex > 0 ? 0 : -1;
            
            // Scroll to first match
            if (this.search.totalMatches > 0) {
                this.scrollToMatch(0);
            }
        },
        
        // Navigate to next search match
        nextMatch() {
            // In edit mode, navigate editor matches only
            if (this.ui.viewMode === 'edit' && this.search.editorMatchCount > 0) {
                this.nextEditorMatch();
                return;
            }
            
            if (this.search.totalMatches === 0) return;
            
            this.search.matchIndex = (this.search.matchIndex + 1) % this.search.totalMatches;
            this.scrollToMatch(this.search.matchIndex);
            
            // In split mode, also navigate the editor to the corresponding match
            if (this.ui.viewMode === 'split' && this.search.editorMatchPositions.length > 0) {
                const editorIndex = Math.min(this.search.matchIndex, this.search.editorMatchPositions.length - 1);
                this.scrollToEditorMatch(editorIndex);
            }
        },
        
        // Navigate to previous search match
        previousMatch() {
            // In edit mode, navigate editor matches only
            if (this.ui.viewMode === 'edit' && this.search.editorMatchCount > 0) {
                this.previousEditorMatch();
                return;
            }
            
            if (this.search.totalMatches === 0) return;
            
            this.search.matchIndex = (this.search.matchIndex - 1 + this.search.totalMatches) % this.search.totalMatches;
            this.scrollToMatch(this.search.matchIndex);
            
            // In split mode, also navigate the editor to the corresponding match
            if (this.ui.viewMode === 'split' && this.search.editorMatchPositions.length > 0) {
                const editorIndex = Math.min(this.search.matchIndex, this.search.editorMatchPositions.length - 1);
                this.scrollToEditorMatch(editorIndex);
            }
        },
        
        // Scroll to a specific match index
        scrollToMatch(index) {
            const preview = NoteCache.domCache.previewContent || document.querySelector('.markdown-preview');
            if (!preview) return;
            
            const allMatches = preview.querySelectorAll('mark.search-highlight');
            if (index < 0 || index >= allMatches.length) return;
            
            // Update styling - make current match prominent (via CSS class)
            allMatches.forEach((mark, i) => {
                mark.classList.toggle('active-match', i === index);
            });
            
            // Scroll to the match
            const targetMatch = allMatches[index];
            const previewContainer = NoteCache.domCache.previewContainer;
            if (previewContainer && targetMatch) {
                const elementTop = targetMatch.offsetTop;
                previewContainer.scrollTop = elementTop - 100; // Scroll with some offset
            }
        },
        
        // Clear preview highlights only (without resetting editor match state)
        clearPreviewHighlights() {
            const preview = NoteCache.domCache.previewContent || document.querySelector('.markdown-preview');
            if (!preview) return;
            
            const highlights = preview.querySelectorAll('mark.search-highlight');
            highlights.forEach(mark => {
                const text = document.createTextNode(mark.textContent);
                mark.parentNode.replaceChild(text, mark);
            });
            
            // Normalize text nodes to merge adjacent text nodes
            preview.normalize();
            
            // Reset preview match counters only
            this.search.totalMatches = 0;
            this.search.matchIndex = -1;
        },
        
        // Clear search highlights
        clearSearchHighlights() {
            const preview = NoteCache.domCache.previewContent || document.querySelector('.markdown-preview');
            if (preview) {
                const highlights = preview.querySelectorAll('mark.search-highlight');
                highlights.forEach(mark => {
                    const text = document.createTextNode(mark.textContent);
                    mark.parentNode.replaceChild(text, mark);
                });
                
                // Normalize text nodes to merge adjacent text nodes
                preview.normalize();
            }
            
            // Reset match counters
            this.search.totalMatches = 0;
            this.search.matchIndex = -1;
            
            // Reset editor match state
            this.search.editorMatchPositions = [];
            this.search.editorMatchCount = 0;
            this.search.currentEditorMatchIndex = -1;
        },
        
        // =====================================================
        // DROPDOWN MENU SYSTEM
        // =====================================================
        
        toggleNewDropdown(event) {
            this.modals.newDropdown.show = true; // Always open (or keep open)
            
            if (event && event.target) {
                const rect = event.target.getBoundingClientRect();
                // Position dropdown next to the clicked element
                let top = rect.bottom + 4; // 4px spacing
                let left = rect.left;
                
                // Keep dropdown on screen
                const dropdownWidth = 200;
                const dropdownHeight = 150;
                if (left + dropdownWidth > window.innerWidth) {
                    left = rect.right - dropdownWidth;
                }
                if (top + dropdownHeight > window.innerHeight) {
                    top = rect.top - dropdownHeight - 4;
                }
                
                this.modals.newDropdown.position = { top, left };
            }
        },
        
        closeDropdown() {
            this.modals.newDropdown.show = false;
            this.modals.newDropdown.targetFolder = null; // Reset folder context
        },
        
        // =====================================================
        // UNIFIED CREATION FUNCTIONS (reusable from anywhere)
        // =====================================================
        
        // Switch to split view (if in preview-only mode) and focus editor for new notes
        focusEditorForNewNote() {
            // Only switch if in preview-only mode - don't disturb edit or split mode
            if (this.ui.viewMode === 'preview') {
                this.ui.viewMode = 'split';
                this.saveViewMode();
            }
            // Focus the editor after a short delay to ensure DOM is updated
            this.$nextTick(() => {
                const editor = document.getElementById('note-editor');
                if (editor) editor.focus();
            });
        },
        
        async createNote(folderPath = null, directPath = null) {
            let notePath;
            
            if (directPath) {
                // Direct path provided (e.g., from wiki link) - skip prompting
                notePath = directPath.endsWith('.md') ? directPath : `${directPath}.md`;
            } else {
                // Use provided folder path, or dropdown target folder context, or homepage folder
                // Note: Check dropdownTargetFolder !== null to distinguish between '' (root) and not set
                let targetFolder;
                if (folderPath !== null) {
                    targetFolder = folderPath;
                } else if (this.modals.newDropdown.targetFolder !== null && this.modals.newDropdown.targetFolder !== undefined) {
                    targetFolder = this.modals.newDropdown.targetFolder; // Can be '' for root or a folder path
                } else {
                    targetFolder = this.homepage.selectedFolder || '';
                }
                this.closeDropdown();
                
                const promptText = targetFolder 
                    ? this.t('notes.prompt_name_in_folder', { folder: targetFolder })
                    : this.t('notes.prompt_name_with_path');
                
                const noteName = await this.showPrompt(promptText);
                if (!noteName) return;
                
                // Validate the name/path (may contain / for paths when no target folder)
                const validation = targetFolder 
                    ? FilenameValidator.validateFilename(noteName)
                    : FilenameValidator.validatePath(noteName);
                
                if (!validation.valid) {
                    this.showAlert(this.getValidationErrorMessage(validation, 'note'));
                    return;
                }
                
                const validatedName = validation.sanitized;
                
                if (targetFolder) {
                    notePath = `${targetFolder}/${validatedName}.md`;
                } else {
                    notePath = validatedName.endsWith('.md') ? validatedName : `${validatedName}.md`;
                }
            }
            
            // CRITICAL: Check if note already exists (applies to both prompt and direct path)
            const existingNote = this.notes.find(note => note.path === notePath);
            if (existingNote) {
                this.showAlert(this.t('notes.already_exists', { name: notePath }));
                return;
            }
            
            try {
                const encodedPath = notePath.split('/').map(segment => encodeURIComponent(segment)).join('/');
                const response = await secureFetch(`/api/notes/${encodedPath}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ content: '' })
                });
                
                if (response.ok) {
                    // Expand parent folder if note is in a subfolder
                    const folderPart = notePath.includes('/') ? notePath.substring(0, notePath.lastIndexOf('/')) : '';
                    if (folderPart) this.folders.expanded.add(folderPart);
                    await this.loadNotes();
                    await this.loadNote(notePath);
                    this.focusEditorForNewNote();
                } else {
                    ErrorHandler.handle('create note', new Error('Server returned error'));
                }
            } catch (error) {
                ErrorHandler.handle('create note', error);
            }
        },
        
        async createFolder(parentPath = null) {
            // Use provided parent path, or dropdown target folder context, or homepage folder
            // Note: Check dropdownTargetFolder !== null to distinguish between '' (root) and not set
            let targetFolder;
            if (parentPath !== null) {
                targetFolder = parentPath;
            } else if (this.modals.newDropdown.targetFolder !== null && this.modals.newDropdown.targetFolder !== undefined) {
                targetFolder = this.modals.newDropdown.targetFolder; // Can be '' for root or a folder path
            } else {
                targetFolder = this.homepage.selectedFolder || '';
            }
            this.closeDropdown();
            
            const promptText = targetFolder 
                ? this.t('folders.prompt_name_in_folder', { folder: targetFolder })
                : this.t('folders.prompt_name_with_path');
            
            const folderName = await this.showPrompt(promptText);
            if (!folderName) return;
            
            // Validate the name/path (may contain / for paths when no target folder)
            const validation = targetFolder 
                ? FilenameValidator.validateFilename(folderName)
                : FilenameValidator.validatePath(folderName);
            
            if (!validation.valid) {
                this.showAlert(this.getValidationErrorMessage(validation, 'folder'));
                return;
            }
            
            const validatedName = validation.sanitized;
            const folderPath = targetFolder ? `${targetFolder}/${validatedName}` : validatedName;
            
            // Check if folder already exists
            const existingFolder = this.folders.all.find(folder => folder === folderPath);
            if (existingFolder) {
                this.showAlert(this.t('folders.already_exists', { name: validatedName }));
                return;
            }
            
            try {
                const response = await secureFetch('/api/folders', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ path: folderPath })
                });
                
                if (response.ok) {
                    if (targetFolder) {
                        this.folders.expanded.add(targetFolder);
                    }
                    this.folders.expanded.add(folderPath);
                    await this.loadNotes();
                    
                    // Navigate to the newly created folder on the homepage
                    this.goToHomepageFolder(folderPath);
                } else {
                    ErrorHandler.handle('create folder', new Error('Server returned error'));
                }
            } catch (error) {
                ErrorHandler.handle('create folder', error);
            }
        },
        
        // Rename a folder
        async renameFolder(folderPath, currentName) {
            const newName = await this.showPrompt(this.t('folders.prompt_rename', { name: currentName }), currentName);
            if (!newName || newName === currentName) return;
            
            // Validate the new name (single segment, no path separators)
            const validation = FilenameValidator.validateFilename(newName);
            if (!validation.valid) {
                this.showAlert(this.getValidationErrorMessage(validation, 'folder'));
                return;
            }
            
            const validatedName = validation.sanitized;
            
            // Calculate new path
            const pathParts = folderPath.split('/');
            pathParts[pathParts.length - 1] = validatedName;
            const newPath = pathParts.join('/');
            
            try {
                const response = await secureFetch('/api/folders/rename', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ 
                        oldPath: folderPath,
                        newPath: newPath
                    })
                });
                
                if (response.ok) {
                    // Update expanded folders state
                    if (this.folders.expanded.has(folderPath)) {
                        this.folders.expanded.delete(folderPath);
                        this.folders.expanded.add(newPath);
                    }
                    
                    // Update favorites that were in the renamed folder
                    const folderPrefix = folderPath + '/';
                    const newFolderPrefix = newPath + '/';
                    const newFavorites = this._favoritesState.list.map(f => {
                        if (f.startsWith(folderPrefix)) {
                            return f.replace(folderPrefix, newFolderPrefix);
                        }
                        return f;
                    });
                    // Check if anything changed
                    if (JSON.stringify(newFavorites) !== JSON.stringify(this._favoritesState.list)) {
                        this._favoritesState.list = newFavorites;
                        this._favoritesState.set = new Set(newFavorites);
                        this.saveFavorites();
                    }
                    
                    // Update current note path if it's in the renamed folder
                    if (this.note.current && this.note.current.startsWith(folderPrefix)) {
                        this.note.current = this.note.current.replace(folderPrefix, newFolderPrefix);
                    }
                    
                    await this.loadNotes();
                } else {
                    ErrorHandler.handle('rename folder', new Error('Server returned error'));
                }
            } catch (error) {
                ErrorHandler.handle('rename folder', error);
            }
        },
        
        // Delete folder
        async deleteFolder(folderPath, folderName) {
            if (!await this.showConfirm(this.t('folders.confirm_delete', { name: folderName }))) return;
            
            try {
                const response = await secureFetch(`/api/folders/${encodeURIComponent(folderPath)}`, {
                    method: 'DELETE',
                    headers: { 'Content-Type': 'application/json' }
                });
                
                if (response.ok) {
                    // Remove from expanded folders
                    this.folders.expanded.delete(folderPath);
                    
                    // Remove any favorites that were in the deleted folder
                    const folderPrefix = folderPath + '/';
                    const newFavorites = this._favoritesState.list.filter(f => !f.startsWith(folderPrefix));
                    if (newFavorites.length !== this._favoritesState.list.length) {
                        this._favoritesState.list = newFavorites;
                        this._favoritesState.set = new Set(newFavorites);
                        this.saveFavorites();
                    }
                    
                    // Clear current note if it was in the deleted folder
                    if (this.note.current && this.note.current.startsWith(folderPrefix)) {
                        this.note.current = '';
                        this.note.content = '';
                        document.title = this.app.name;
                    }
                    
                    await this.loadNotes();
                } else {
                    ErrorHandler.handle('delete folder', new Error('Server returned error'));
                }
            } catch (error) {
                ErrorHandler.handle('delete folder', error);
            }
        },
        
        // Auto-save with debounce
        autoSave() {
            if (this.note.saveTimeout) {
                clearTimeout(this.note.saveTimeout);
            }
            
            this.note.dirty = true;
            this.note.lastSaved = false;
            
            // Push to undo history (but not during undo/redo operations)
            if (!this.history.isUndoRedo) {
                this.pushToHistory();
            }
            
            // Calculate stats in real-time if enabled
            if (this.stats.enabled) {
                this.calculateStats();
            }
            
            // Parse metadata in real-time
            this.parseMetadata();
            
            // Update outline (TOC) in real-time
            this.extractOutline(this.note.content);
            
            this.note.saveTimeout = setTimeout(() => {
                // Commit to undo history when autosave triggers (same debounce timing)
                if (this.history.hasPendingChanges) {
                    this.commitToHistory();
                }
                this.saveNote();
            }, CONFIG.AUTOSAVE_DELAY);
        },
        
        // Mark that we have pending changes (called on each keystroke)
        pushToHistory() {
            this.history.hasPendingChanges = true;
        },
        
        // Immediately commit pending changes to history (call before undo/redo)
        flushHistory() {
            if (this.history.hasPendingChanges) {
                this.commitToHistory();
            }
        },
        
        // Actually commit to undo history (internal)
        commitToHistory() {
            const editor = document.getElementById('note-editor');
            const cursorPos = editor ? editor.selectionStart : 0;
            
            // Only push if content actually changed from last history entry
            if (this.history.undo.length > 0 && 
                this.history.undo[this.history.undo.length - 1].content === this.note.content) {
                this.history.hasPendingChanges = false;
                return;
            }
            
            this.history.undo.push({ content: this.note.content, cursorPos });
            
            // Limit history size
            if (this.history.undo.length > this.history.maxSize) {
                this.history.undo.shift();
            }
            
            // Clear redo history when new change is made
            this.history.redo = [];
            this.history.hasPendingChanges = false;
        },
        
        // Undo last change
        undo() {
            if (!this.note.current) return;
            
            // Flush any pending history changes first (so we don't lose unsaved edits)
            this.flushHistory();
            
            if (this.history.undo.length <= 1) return;
            
            const editor = document.getElementById('note-editor');
            
            // Pop current state to redo history
            const currentState = this.history.undo.pop();
            this.history.redo.push(currentState);
            
            // Get previous state
            const previousState = this.history.undo[this.history.undo.length - 1];
            
            // Apply previous state
            this.history.isUndoRedo = true;
            this.note.content = previousState.content;
            
            // Recalculate stats with new content
            if (this.stats.enabled) {
                this.calculateStats();
            }
            
            // Restore cursor position from the state we're going back to
            this.$nextTick(() => {
                this.saveNote();
                this.history.isUndoRedo = false;
                if (editor) {
                    setTimeout(() => {
                        const newPos = Math.min(previousState.cursorPos, this.note.content.length);
                        editor.setSelectionRange(newPos, newPos);
                        editor.focus();
                    }, 0);
                }
            });
        },
        
        // Redo last undone change
        redo() {
            if (!this.note.current) return;
            
            // Flush any pending history changes first
            this.flushHistory();
            
            if (this.history.redo.length === 0) return;
            
            const editor = document.getElementById('note-editor');
            
            // Pop from redo history
            const nextState = this.history.redo.pop();
            
            // Push to undo history
            this.history.undo.push(nextState);
            
            // Apply next state
            this.history.isUndoRedo = true;
            this.note.content = nextState.content;
            
            // Recalculate stats with new content
            if (this.stats.enabled) {
                this.calculateStats();
            }
            
            // Restore cursor position from the state we're going forward to
            this.$nextTick(() => {
                this.saveNote();
                this.history.isUndoRedo = false;
                if (editor) {
                    setTimeout(() => {
                        const newPos = Math.min(nextState.cursorPos, this.note.content.length);
                        editor.setSelectionRange(newPos, newPos);
                        editor.focus();
                    }, 0);
                }
            });
        },
        
        // Markdown formatting helpers
        wrapSelection(before, after, placeholder) {
            const editor = document.getElementById('note-editor');
            if (!editor) return;
            
            const start = editor.selectionStart;
            const end = editor.selectionEnd;
            const selectedText = this.note.content.substring(start, end);
            const textToWrap = selectedText || placeholder;
            
            // Build the new text
            const newText = before + textToWrap + after;
            
            // Update content
            this.note.content = this.note.content.substring(0, start) + newText + this.note.content.substring(end);
            
            // Set cursor position (select the wrapped text or placeholder)
            this.$nextTick(() => {
                if (selectedText) {
                    // If text was selected, keep it selected (inside the wrapper)
                    editor.setSelectionRange(start + before.length, start + before.length + selectedText.length);
                } else {
                    // If no text selected, select the placeholder
                    editor.setSelectionRange(start + before.length, start + before.length + placeholder.length);
                }
                editor.focus();
            });
            
            // Trigger autosave
            this.autoSave();
        },
        
        insertLink() {
            const editor = document.getElementById('note-editor');
            if (!editor) return;
            
            const start = editor.selectionStart;
            const end = editor.selectionEnd;
            const selectedText = this.note.content.substring(start, end);
            
            // If text is selected, use it as link text; otherwise use placeholder
            const linkText = selectedText || 'link text';
            const linkUrl = 'url';
            
            // Build the markdown link
            const newText = `[${linkText}](${linkUrl})`;
            
            // Update content
            this.note.content = this.note.content.substring(0, start) + newText + this.note.content.substring(end);
            
            // Set cursor position to select the URL part for easy editing
            this.$nextTick(() => {
                const urlStart = start + linkText.length + 3; // After "[linkText]("
                const urlEnd = urlStart + linkUrl.length;
                editor.setSelectionRange(urlStart, urlEnd);
                editor.focus();
            });
            
            // Trigger autosave
            this.autoSave();
        },
        
        // Insert a markdown table placeholder
        insertTable() {
            const editor = document.getElementById('note-editor');
            if (!editor) return;
            
            const cursorPos = editor.selectionStart;
            
            // Basic 3x3 table placeholder
            const table = `| Header 1 | Header 2 | Header 3 |
|----------|----------|----------|
| Cell 1   | Cell 2   | Cell 3   |
| Cell 4   | Cell 5   | Cell 6   |
`;
            
            // Add newline before if not at start of line
            const textBefore = this.note.content.substring(0, cursorPos);
            const needsNewlineBefore = textBefore.length > 0 && !textBefore.endsWith('\n');
            const prefix = needsNewlineBefore ? '\n\n' : '';
            
            // Insert the table
            this.note.content = textBefore + prefix + table + this.note.content.substring(cursorPos);
            
            // Position cursor at first header for easy editing
            this.$nextTick(() => {
                const newPos = cursorPos + prefix.length + 2; // After "| "
                editor.setSelectionRange(newPos, newPos + 8); // Select "Header 1"
                editor.focus();
            });
            
            // Trigger autosave
            this.autoSave();
        },
        
        // Format selected text or insert formatting at cursor
        formatText(type) {
            // Simple wrap cases - reuse wrapSelection()
            const wrapFormats = {
                'bold': ['**', '**', 'bold'],
                'italic': ['*', '*', 'italic'],
                'strikethrough': ['~~', '~~', 'strikethrough'],
                'code': ['`', '`', 'code']
            };
            
            if (wrapFormats[type]) {
                const [before, after, placeholder] = wrapFormats[type];
                this.wrapSelection(before, after, placeholder);
                return;
            }
            
            // Special cases that need custom handling
            switch (type) {
                case 'heading':
                    this.insertLinePrefix('## ', 'Heading');
                    break;
                case 'quote':
                    this.insertLinePrefix('> ', 'quote');
                    break;
                case 'bullet':
                    this.insertLinePrefix('- ', 'item');
                    break;
                case 'numbered':
                    this.insertLinePrefix('1. ', 'item');
                    break;
                case 'checkbox':
                    this.insertLinePrefix('- [ ] ', 'task');
                    break;
                case 'link':
                    this.insertLink();
                    break;
                case 'image':
                    this.wrapSelection('![', '](image-url)', 'alt text');
                    break;
                case 'codeblock':
                    this.wrapSelection('```\n', '\n```', 'code');
                    break;
                case 'table':
                    this.insertTable();
                    break;
            }
        },
        
        // Insert a line prefix (for headings, lists, quotes)
        insertLinePrefix(prefix, placeholder) {
            const editor = document.getElementById('note-editor');
            if (!editor) return;
            
            const start = editor.selectionStart;
            const end = editor.selectionEnd;
            const selectedText = this.note.content.substring(start, end);
            const beforeText = this.note.content.substring(0, start);
            const afterText = this.note.content.substring(end);
            
            // Check if at start of line
            const atLineStart = beforeText.endsWith('\n') || beforeText === '';
            const newline = atLineStart ? '' : '\n';
            
            let replacement;
            if (selectedText) {
                // Prefix each line of selection
                replacement = newline + selectedText.split('\n').map((line, i) => {
                    // For numbered lists, increment the number
                    if (prefix === '1. ') return `${i + 1}. ${line}`;
                    return prefix + line;
                }).join('\n');
            } else {
                replacement = newline + prefix + placeholder;
            }
            
            this.note.content = beforeText + replacement + afterText;
            
            this.$nextTick(() => {
                if (selectedText) {
                    editor.setSelectionRange(start + newline.length, start + replacement.length);
                } else {
                    const placeholderStart = start + newline.length + prefix.length;
                    editor.setSelectionRange(placeholderStart, placeholderStart + placeholder.length);
                }
                editor.focus();
            });
            
            this.autoSave();
        },
        
        // Save current note with retry mechanism
        async saveNote(retryCount = 0, saveId = null) {
            if (!this.note.current) return;
            
            // Only set isSaving on first attempt, not during retries
            if (retryCount === 0) {
                this.note.isSaving = true;
                // Generate a unique save ID for this save operation
                this._currentSaveId = Date.now() + Math.random();
                this._savingNotePath = this.note.current;
            }
            
            // Check if this save operation has been superseded by a newer one
            if (saveId !== null && saveId !== this._currentSaveId) {
                console.log('Save cancelled: newer save operation in progress');
                return;
            }
            
            // Validate that we're still on the same note during retries
            if (retryCount > 0 && this._savingNotePath !== this.note.current) {
                console.log('Save cancelled: user switched to a different note');
                this.note.isSaving = false;
                this._savingNotePath = null;
                this._currentSaveId = null;
                return;
            }
            
            // Always use the current content (not cached from first attempt)
            const contentToSave = this.note.content;
            const notePath = this.note.current;
            const currentSaveId = this._currentSaveId;
            
            try {
                const encodedPath = notePath.split('/').map(segment => {
                    try {
                        return encodeURIComponent(decodeURIComponent(segment));
                    } catch (e) {
                        return encodeURIComponent(segment);
                    }
                }).join('/');
                const response = await secureFetch(`/api/notes/${encodedPath}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        content: contentToSave,
                        modified: this.note.modified || ''
                    })
                });
                
                if (response.ok) {
                    const data = await response.json();
                    this.note.lastSaved = true;
                    
                    // Clear localStorage backup on successful save
                    this.clearDraft(notePath);
                    
                    // Update using server-authoritative mtime
                    this.note.modified = data.modified || '';
                    this.note.dirty = false;
                    
                    // Update list cache entry
                    const note = this.notes.find(n => n.path === notePath);
                    if (note) {
                        note.modified = data.modified || '';
                        note.size = new Blob([contentToSave]).size;
                        
                        // Parse tags from content
                        note.tags = this.parseTagsFromContent(contentToSave);
                    }
                    
                    // Reload tags immediately to update sidebar counts
                    this.loadTags();
                    
                    // Rebuild folder tree if tag filters are active
                    if (this.tags.selected.length > 0) {
                        this.buildFolderTree();
                    }
                    
                    // Hide "saved" indicator
                    setTimeout(() => {
                        this.note.lastSaved = false;
                    }, CONFIG.SAVE_INDICATOR_DURATION);
                    
                    // Clear saving state on success
                    this.note.isSaving = false;
                    this._savingNotePath = null;
                    this._currentSaveId = null;
                } else if (response.status === 409) {
                    // Conflict: note was modified by another source
                    this.note.isSaving = false;
                    this._savingNotePath = null;
                    this._currentSaveId = null;
                    const data = await response.json().catch(() => ({}));
                    this.showConflictBanner(notePath, data.modified || '');
                } else {
                    throw new Error(`Server returned error: ${response.status}`);
                }
            } catch (error) {
                // Save to localStorage as backup before retrying
                this.writeDraft(notePath, contentToSave);
                
                // Retry with exponential backoff
                if (retryCount < CONFIG.MAX_SAVE_RETRIES) {
                    const delay = CONFIG.SAVE_RETRY_BASE_DELAY * Math.pow(2, retryCount);
                    console.log(`Save failed, retrying in ${delay}ms (attempt ${retryCount + 1}/${CONFIG.MAX_SAVE_RETRIES})`);
                    
                    await new Promise(resolve => setTimeout(resolve, delay));
                    // Pass the saveId to validate this is still the active save
                    return this.saveNote(retryCount + 1, currentSaveId);
                }
                
                // All retries exhausted
                console.error('Save failed after all retries:', error);
                this.note.isSaving = false;
                this._savingNotePath = null;
                this._currentSaveId = null;
                this.showSaveErrorToast();
            }
            // No finally block - isSaving is managed explicitly above
        },

        // Write draft to localStorage (multi-slot, per-path)
        writeDraft(path, content) {
            if (!path || !content) return;
            try {
                localStorage.setItem('gonote_draft:' + path, JSON.stringify({
                    content: content,
                    timestamp: Date.now()
                }));
            } catch (e) {
                if (e.name === 'QuotaExceededError' || e.code === 22) {
                    // Quota exceeded: remove 10 oldest drafts and retry
                    const drafts = this.listDrafts().sort((a, b) => a.timestamp - b.timestamp);
                    drafts.slice(0, 10).forEach(d => {
                        try { localStorage.removeItem('gonote_draft:' + d.path); } catch (_) {}
                    });
                    try {
                        localStorage.setItem('gonote_draft:' + path, JSON.stringify({
                            content: content,
                            timestamp: Date.now()
                        }));
                    } catch (_) {}
                }
            }
        },

        // Clear draft for a specific path
        clearDraft(path) {
            if (!path) return;
            try {
                localStorage.removeItem('gonote_draft:' + path);
            } catch (e) {
                console.warn('Failed to clear draft:', e);
            }
        },

        // List all drafts, removing those older than 7 days
        listDrafts() {
            const drafts = [];
            const now = Date.now();
            const maxAge = 7 * 24 * 60 * 60 * 1000;
            try {
                for (let i = 0; i < localStorage.length; i++) {
                    const key = localStorage.key(i);
                    if (key && key.startsWith('gonote_draft:')) {
                        const data = JSON.parse(localStorage.getItem(key));
                        if (now - data.timestamp > maxAge) {
                            localStorage.removeItem(key);
                            continue;
                        }
                        drafts.push({
                            path: key.slice('gonote_draft:'.length),
                            content: data.content,
                            timestamp: data.timestamp
                        });
                    }
                }
            } catch (e) {
                console.warn('Failed to list drafts:', e);
            }
            return drafts;
        },

        // Check and restore drafts on startup
        checkAndRestoreDrafts() {
            const drafts = this.listDrafts();
            if (drafts.length > 0) {
                this.showDraftsRestoreModal(drafts);
            }
        },

        // Show draft restore modal (multi-draft list)
        showDraftsRestoreModal(drafts) {
            const modal = document.createElement('div');
            modal.id = 'drafts-restore-modal';
            modal.style.cssText = `
                position: fixed; top: 0; left: 0; right: 0; bottom: 0;
                background: rgba(0,0,0,0.5); display: flex;
                align-items: center; justify-content: center; z-index: 10001;
            `;

            const container = document.createElement('div');
            container.style.cssText = `
                background: var(--bg-primary, #1e1e1e);
                color: var(--text-primary, #e0e0e0);
                padding: 24px; border-radius: 12px;
                max-width: 480px; width: 90%;
                box-shadow: 0 8px 32px rgba(0,0,0,0.4);
                max-height: 80vh; overflow-y: auto;
            `;

            let itemsHtml = '';
            drafts.sort((a, b) => b.timestamp - a.timestamp).forEach(d => {
                const timeAgo = this.formatTimeAgo(d.timestamp);
                const noteName = d.path.split('/').pop().replace('.md', '');
                itemsHtml += `
                    <div class="draft-item" style="
                        display: flex; align-items: center; justify-content: space-between;
                        padding: 10px 0; border-bottom: 1px solid var(--border-color, #333);
                    ">
                        <div style="flex: 1; min-width: 0;">
                            <div style="font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">📄 ${noteName}</div>
                            <div style="font-size: 12px; color: var(--text-secondary, #888);">${timeAgo}</div>
                        </div>
                        <div style="display: flex; gap: 8px; flex-shrink: 0; margin-left: 12px;">
                            <button class="draft-restore-btn" data-path="${d.path}" style="
                                padding: 6px 12px; font-size: 12px;
                                border: none; background: #3b82f6; color: white;
                                border-radius: 4px; cursor: pointer;
                            ">${this.t('common.restore') || 'Restore'}</button>
                            <button class="draft-discard-btn" data-path="${d.path}" style="
                                padding: 6px 12px; font-size: 12px;
                                border: 1px solid var(--border-color, #444); background: transparent;
                                color: var(--text-primary, #e0e0e0); border-radius: 4px; cursor: pointer;
                            ">${this.t('common.discard') || 'Discard'}</button>
                        </div>
                    </div>
                `;
            });

            const count = drafts.length;
            container.innerHTML = `
                <h3 style="margin: 0 0 16px 0; font-size: 18px;">
                    ${this.t('notes.unsaved_drafts') || 'Unsaved Drafts'} (${count})
                </h3>
                <div style="margin-bottom: 16px;">${itemsHtml}</div>
                <div style="display: flex; gap: 8px; justify-content: flex-end; padding-top: 8px; border-top: 1px solid var(--border-color, #333);">
                    <button id="drafts-discard-all-btn" style="
                        padding: 8px 16px; font-size: 13px;
                        border: 1px solid var(--border-color, #444); background: transparent;
                        color: var(--text-primary, #e0e0e0); border-radius: 6px; cursor: pointer;
                    ">${this.t('notes.discard_all') || 'Discard All'}</button>
                    <button id="drafts-remind-later-btn" style="
                        padding: 8px 16px; font-size: 13px;
                        border: 1px solid var(--border-color, #444); background: transparent;
                        color: var(--text-primary, #e0e0e0); border-radius: 6px; cursor: pointer;
                    ">${this.t('notes.remind_later') || 'Remind Later'}</button>
                </div>
            `;

            modal.appendChild(container);
            document.body.appendChild(modal);

            // Handle individual restore
            container.querySelectorAll('.draft-restore-btn').forEach(btn => {
                btn.onclick = () => {
                    const path = btn.dataset.path;
                    const draft = drafts.find(d => d.path === path);
                    if (draft) this.restoreSingleDraft(draft, drafts);
                };
            });

            // Handle individual discard
            container.querySelectorAll('.draft-discard-btn').forEach(btn => {
                btn.onclick = () => {
                    this.clearDraft(btn.dataset.path);
                    btn.closest('.draft-item').remove();
                    if (container.querySelectorAll('.draft-item').length === 0) {
                        document.body.removeChild(modal);
                    }
                };
            });

            // Discard all
            document.getElementById('drafts-discard-all-btn').onclick = () => {
                drafts.forEach(d => this.clearDraft(d.path));
                document.body.removeChild(modal);
            };

            // Remind later - just close
            document.getElementById('drafts-remind-later-btn').onclick = () => {
                document.body.removeChild(modal);
            };

            // Close on background click
            modal.onclick = (e) => {
                if (e.target === modal) document.body.removeChild(modal);
            };
        },

        // Restore a single draft (with conflict check)
        restoreSingleDraft(draft, drafts) {
            this.loadNote(draft.path).then(() => {
                const serverMtime = this.note.modified;
                const draftTime = draft.timestamp;
                const serverTime = serverMtime ? Date.parse(serverMtime) : 0;

                if (serverTime > draftTime) {
                    // Server has newer content: prompt conflict resolution
                    this.showDraftConflictModal(draft, drafts);
                } else {
                    // Draft is newer or server has no timestamp: restore directly
                    this.note.content = draft.content;
                    this.note.dirty = true;
                    this.clearDraft(draft.path);
                    // Close the restore modal
                    const modal = document.getElementById('drafts-restore-modal');
                    if (modal) document.body.removeChild(modal);
                }
            });
        },

        // Show draft vs server conflict modal
        showDraftConflictModal(draft, drafts) {
            const modal = document.createElement('div');
            modal.id = 'draft-conflict-modal';
            modal.style.cssText = `
                position: fixed; top: 0; left: 0; right: 0; bottom: 0;
                background: rgba(0,0,0,0.5); display: flex;
                align-items: center; justify-content: center; z-index: 10002;
            `;

            const timeAgo = this.formatTimeAgo(draft.timestamp);
            const serverTimeStr = this.note.modified ? new Date(this.note.modified).toLocaleString() : 'unknown';

            const container = document.createElement('div');
            container.style.cssText = `
                background: var(--bg-primary, #1e1e1e);
                color: var(--text-primary, #e0e0e0);
                padding: 24px; border-radius: 12px; max-width: 420px;
                box-shadow: 0 8px 32px rgba(0,0,0,0.4);
            `;
            container.innerHTML = `
                <h3 style="margin: 0 0 12px 0; font-size: 16px;">⚠ ${this.t('notes.conflict_draft_older') || 'Server has newer version'}</h3>
                <p style="margin: 0 0 8px 0; font-size: 13px; color: var(--text-secondary, #888);">
                    ${this.t('notes.draft_saved_at') || 'Draft saved'}: ${timeAgo}<br>
                    ${this.t('notes.server_version_at') || 'Server version'}: ${serverTimeStr}
                </p>
                <p style="margin: 0 0 16px 0; font-size: 13px; color: var(--text-secondary, #888);">
                    ${this.t('notes.choose_version') || 'Choose which version to keep:'}
                </p>
                <div style="display: flex; gap: 8px; justify-content: flex-end;">
                    <button id="draft-load-draft-btn" style="
                        padding: 8px 16px; font-size: 13px;
                        border: none; background: #3b82f6; color: white; border-radius: 6px; cursor: pointer;
                    ">${this.t('notes.load_draft') || 'Load Draft'}</button>
                    <button id="draft-load-server-btn" style="
                        padding: 8px 16px; font-size: 13px;
                        border: 1px solid var(--border-color, #444); background: transparent;
                        color: var(--text-primary, #e0e0e0); border-radius: 6px; cursor: pointer;
                    ">${this.t('notes.load_server_version') || 'Load Server Version'}</button>
                </div>
            `;

            modal.appendChild(container);
            document.body.appendChild(modal);

            document.getElementById('draft-load-draft-btn').onclick = () => {
                this.note.content = draft.content;
                this.note.dirty = true;
                this.clearDraft(draft.path);
                document.body.removeChild(modal);
                const restoreModal = document.getElementById('drafts-restore-modal');
                if (restoreModal) document.body.removeChild(restoreModal);
            };

            document.getElementById('draft-load-server-btn').onclick = () => {
                this.loadNote(draft.path);
                this.clearDraft(draft.path);
                document.body.removeChild(modal);
                const restoreModal = document.getElementById('drafts-restore-modal');
                if (restoreModal) document.body.removeChild(restoreModal);
            };

            modal.onclick = (e) => {
                if (e.target === modal) document.body.removeChild(modal);
            };
        },

        // Show conflict banner when server returns 409
        showConflictBanner(notePath, serverMtime) {
            // Remove existing banner if any
            const existing = document.getElementById('conflict-banner');
            if (existing) existing.remove();

            const banner = document.createElement('div');
            banner.id = 'conflict-banner';
            banner.style.cssText = `
                position: fixed; top: 0; left: 0; right: 0;
                background: #f59e0b; color: #1e1e1e;
                padding: 12px 20px; display: flex;
                align-items: center; justify-content: center; gap: 12px;
                z-index: 10000; font-size: 14px;
                box-shadow: 0 2px 8px rgba(0,0,0,0.2);
            `;

            const serverTimeStr = serverMtime ? new Date(serverMtime).toLocaleString() : '';

            banner.innerHTML = `
                <span>⚠ ${this.t('notes.conflict_external_modified') || 'Note modified by another source'}${serverTimeStr ? ' at ' + serverTimeStr : ''}</span>
                <button id="conflict-load-server-btn" style="
                    padding: 6px 12px; border: none; background: #1e1e1e;
                    color: #fff; border-radius: 4px; cursor: pointer; font-size: 12px;
                ">${this.t('notes.load_server_version') || 'Load Server Version'}</button>
                <button id="conflict-keep-mine-btn" style="
                    padding: 6px 12px; border: 1px solid #1e1e1e; background: transparent;
                    color: #1e1e1e; border-radius: 4px; cursor: pointer; font-size: 12px;
                ">${this.t('notes.keep_my_version') || 'Keep My Version (Overwrite)'}</button>
            `;

            document.body.insertBefore(banner, document.body.firstChild);

            document.getElementById('conflict-load-server-btn').onclick = () => {
                this.loadNote(notePath);
                this.clearDraft(notePath);
                banner.remove();
            };

            document.getElementById('conflict-keep-mine-btn').onclick = () => {
                // Force save without mtime check (clear modified and re-POST)
                this.note.modified = '';
                this.clearDraft(notePath);
                banner.remove();
                this.saveNote();
            };

            // Auto-dismiss after 30 seconds
            setTimeout(() => {
                if (banner.parentNode) banner.remove();
            }, 30000);
},

        // Show save error toast (non-blocking notification)
        showSaveErrorToast() {
            let toast = document.getElementById('save-error-toast');
            if (!toast) {
                toast = document.createElement('div');
                toast.id = 'save-error-toast';
                toast.style.cssText = `
                    position: fixed; bottom: 20px; right: 20px;
                    background: #dc2626; color: white;
                    padding: 12px 20px; border-radius: 8px;
                    box-shadow: 0 4px 12px rgba(0,0,0,0.3);
                    z-index: 10000; font-size: 14px; max-width: 300px;
                    opacity: 0; transform: translateY(20px);
                    transition: all 0.3s ease;
                `;
                document.body.appendChild(toast);
            }
            toast.textContent = this.t('notes.save_failed') || 'Save failed. Content backed up locally.';
            toast.style.opacity = '1';
            toast.style.transform = 'translateY(0)';
            setTimeout(() => {
                toast.style.opacity = '0';
                toast.style.transform = 'translateY(20px)';
            }, 5000);
        },

        // Format timestamp to human readable time ago
        formatTimeAgo(timestamp) {
            const seconds = Math.floor((Date.now() - timestamp) / 1000);
            
            if (seconds < 60) return this.t('time.just_now') || 'just now';
            if (seconds < 3600) {
                const mins = Math.floor(seconds / 60);
                return this.t('time.minutes_ago', { n: mins }) || `${mins} minutes ago`;
            }
            if (seconds < 86400) {
                const hours = Math.floor(seconds / 3600);
                return this.t('time.hours_ago', { n: hours }) || `${hours} hours ago`;
            }
            const days = Math.floor(seconds / 86400);
            return this.t('time.days_ago', { n: days }) || `${days} days ago`;
        },
        
        // Rename current note
        async renameNote() {
            if (!this.note.current) return;
            
            const oldPath = this.note.current;
            const newName = this.note.name.trim();
            
            if (!newName) {
                this.showAlert(this.t('notes.empty_name'));
                return;
            }
            
            // Validate the new name (single segment, no path separators)
            const validation = FilenameValidator.validateFilename(newName);
            if (!validation.valid) {
                this.showAlert(this.getValidationErrorMessage(validation, 'note'));
                // Reset the name in the UI
                this.note.name = this.decodeNoteName(oldPath.split('/').pop().replace('.md', ''));
                return;
            }
            
            const validatedName = validation.sanitized;
            const folder = oldPath.split('/').slice(0, -1).join('/');
            const newPath = folder ? `${folder}/${validatedName}.md` : `${validatedName}.md`;
            
            if (oldPath === newPath) return;
            
            // Check if a note with the new name already exists
            const existingNote = this.notes.find(n => n.path.toLowerCase() === newPath.toLowerCase());
            if (existingNote) {
                this.showAlert(this.t('notes.already_exists', { name: validatedName }));
                // Reset the name in the UI
                this.note.name = this.decodeNoteName(oldPath.split('/').pop().replace('.md', ''));
                return;
            }
            
            // Create new note with same content
            try {
                const encodedNewPath = newPath.split('/').map(segment => encodeURIComponent(segment)).join('/');
                const response = await secureFetch(`/api/notes/${encodedNewPath}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ content: this.note.content })
                });
                
                if (response.ok) {
                    // Delete old note
                    const encodedOldPath = oldPath.split('/').map(segment => encodeURIComponent(segment)).join('/');
                    await secureFetch(`/api/notes/${encodedOldPath}`, { method: 'DELETE' });
                    
                    // Update favorites if the renamed note was favorited
                    if (this._favoritesState.set.has(oldPath)) {
                        const newFavorites = this._favoritesState.list.map(f => f === oldPath ? newPath : f);
                        this._favoritesState.list = newFavorites;
                        this._favoritesState.set = new Set(newFavorites);
                        this.saveFavorites();
                    }
                    
                    this.note.current = newPath;
                    await this.loadNotes();
                } else {
                    ErrorHandler.handle('rename note', new Error('Server returned error'));
                }
            } catch (error) {
                ErrorHandler.handle('rename note', error);
            }
        },
        
        // Delete current note
        async deleteCurrentNote() {
            if (!this.note.current) return;
            
            // Just call deleteNote with current note details
            await this.deleteNote(this.note.current, this.note.name);
        },
        
        // Delete any note from sidebar
        async deleteNote(notePath, noteName) {
            if (!await this.showConfirm(this.t('notes.confirm_delete', { name: noteName }))) return;
            
            try {
                const encodedPath = notePath.split('/').map(segment => encodeURIComponent(segment)).join('/');
                const response = await secureFetch(`/api/notes/${encodedPath}`, {
                    method: 'DELETE'
                });
                
                if (response.ok) {
                    // Clear draft if any
                    this.clearDraft(notePath);
                    
                    // Remove from favorites if it was favorited
                    if (this._favoritesState.set.has(notePath)) {
                        const newFavorites = this._favoritesState.list.filter(f => f !== notePath);
                        this._favoritesState.list = newFavorites;
                        this._favoritesState.set = new Set(newFavorites);
                        this.saveFavorites();
                    }
                    
                    // If the deleted note is currently open, clear it
                    if (this.note.current === notePath) {
                        this.note.current = '';
                        this.note.content = '';
                        this.note.name = '';
                        NoteCache.lastRenderedContent = ''; // Clear render cache
                        NoteCache.cachedRenderedHTML = '';
                        document.title = this.app.name;
                        // Redirect to root
                        window.history.replaceState({}, '', '/');
                    }
                    
                    await this.loadNotes();
                } else {
                    ErrorHandler.handle('delete note', new Error('Server returned error'));
                }
            } catch (error) {
                ErrorHandler.handle('delete note', error);
            }
        },
        
        // Search notes
        debouncedSearchNotes() {
            if (this.search.debounceTimeout) {
                clearTimeout(this.search.debounceTimeout);
            }

            const hasTextSearch = this.search.query.trim().length > 0;
            if (!hasTextSearch) {
                this.search.isSearching = false;
                this.searchNotes();
                return;
            }

            this.search.isSearching = true;
            this.search.results = [];
            this.search.page = 1; // Reset to first page on new search

            this.search.debounceTimeout = setTimeout(() => {
                this.searchNotes();
            }, CONFIG.SEARCH_DEBOUNCE_DELAY);
        },

        // Search notes by text (calls unified filter logic)
        async searchNotes() {
            await this.applyFilters();
        },
        
        // Navigate to next search page
        nextPage() {
            if (this.search.page < this.search.totalPages) {
                this.search.page++;
                this.searchNotes();
            }
        },
        
        // Navigate to previous search page
        prevPage() {
            if (this.search.page > 1) {
                this.search.page--;
                this.searchNotes();
            }
        },
        
        // Go to specific search page
        goToSearchPage(page) {
            if (page >= 1 && page <= this.search.totalPages && page !== this.search.page) {
                this.search.page = page;
                this.searchNotes();
            }
        },
        
        // Trigger MathJax typesetting after DOM update
        typesetMath() {
            if (typeof MathJax !== 'undefined' && MathJax.typesetPromise) {
                // Use a small delay to ensure DOM is updated
                setTimeout(() => {
                    const previewContent = NoteCache.domCache.previewContent || document.querySelector('.markdown-preview');
                    if (previewContent) {
                        MathJax.typesetPromise([previewContent]).catch((err) => {
                            console.error('MathJax typesetting failed:', err);
                        });
                    }
                }, 10);
            }
        },
        
        // Render Mermaid diagrams
        async renderMermaid() {
            if (typeof window.mermaid === 'undefined') {
                console.warn('Mermaid not loaded yet');
                return;
            }
            
            // Use requestAnimationFrame for better performance than setTimeout
            requestAnimationFrame(async () => {
                const previewContent = NoteCache.domCache.previewContent || document.querySelector('.markdown-preview');
                if (!previewContent) return;
                
                // Get the appropriate theme based on current app theme
                const themeType = this.getThemeType();
                const mermaidTheme = themeType === 'light' ? 'default' : 'dark';
                
                // Only reinitialize if theme changed (performance optimization)
                if (this.graph.lastTheme !== mermaidTheme) {
                    window.mermaid.initialize({ 
                        startOnLoad: false,
                        theme: mermaidTheme,
                        securityLevel: 'strict', // Use strict for better security
                        fontFamily: 'inherit',
                        // v11 changed useMaxWidth defaults - restore responsive behavior
                        flowchart: { useMaxWidth: true },
                        sequence: { useMaxWidth: true },
                        gantt: { useMaxWidth: true },
                        journey: { useMaxWidth: true },
                        timeline: { useMaxWidth: true },
                        class: { useMaxWidth: true },
                        state: { useMaxWidth: true },
                        er: { useMaxWidth: true },
                        pie: { useMaxWidth: true },
                        quadrantChart: { useMaxWidth: true },
                        requirement: { useMaxWidth: true },
                        mindmap: { useMaxWidth: true },
                        gitGraph: { useMaxWidth: true }
                    });
                    this.graph.lastTheme = mermaidTheme;
                }
                
                // Find all code blocks with language 'mermaid'
                const mermaidBlocks = previewContent.querySelectorAll('pre code.language-mermaid');
                
                // Early return if no diagrams to render
                if (mermaidBlocks.length === 0) return;
                
                for (let i = 0; i < mermaidBlocks.length; i++) {
                    const block = mermaidBlocks[i];
                    const pre = block.parentElement;
                    
                    // Skip if already rendered (performance optimization)
                    if (pre.querySelector('.mermaid-rendered')) continue;
                    
                    try {
                        const code = block.textContent;
                        const id = `mermaid-diagram-${Date.now()}-${i}`;
                        
                        // Render the diagram
                        const { svg } = await window.mermaid.render(id, code);
                        
                        // Create a container for the rendered diagram
                        const container = document.createElement('div');
                        container.className = 'mermaid-rendered';
                        container.style.cssText = 'background-color: transparent; padding: 20px; text-align: center; overflow-x: auto;';
                        container.innerHTML = svg;
                        // Store original code for theme re-rendering
                        container.dataset.originalCode = code;
                        
                        // Replace the code block with the rendered diagram
                        pre.parentElement.replaceChild(container, pre);
                    } catch (error) {
                        console.error('Mermaid rendering error:', error);
                        // Add error indicator to the code block
                        const errorMsg = document.createElement('div');
                        errorMsg.style.cssText = 'color: var(--error); padding: 10px; border-left: 3px solid var(--error); margin-top: 10px;';
                        errorMsg.textContent = `⚠️ Mermaid diagram error: ${error.message}`;
                        pre.parentElement.insertBefore(errorMsg, pre.nextSibling);
                    }
                }
            });
        },
        
        // Get current theme type (light or dark)
        // Returns: 'light' or 'dark'
        // Used by features that need to adapt to theme brightness (e.g., Mermaid diagrams, Chart.js)
        getThemeType() {
            // Handle system theme
            if (this.theme.current === 'system') {
                const isDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
                return isDark ? 'dark' : 'light';
            }
            
            // Try to get theme type from loaded theme metadata
            const currentThemeData = this.theme.available.find(t => t.id === this.theme.current);
            if (currentThemeData && currentThemeData.type) {
                // Use metadata from theme file (light or dark)
                return currentThemeData.type; // Already 'light' or 'dark'
            }
            
            // Backward compatibility: fallback to hardcoded map if metadata not available
            const fallbackMap = {
                'light': 'light',
                'vs-blue': 'light'
            };
            
            return fallbackMap[this.theme.current] || 'dark';
        },
        
        
        // Computed property for rendered markdown
        get renderedMarkdown() {
            if (!this.note.content) return '<p style="color: var(--text-tertiary);">Nothing to preview yet...</p>';
            
            // Performance: Return cached HTML if content hasn't changed
            if (this.note.content === NoteCache.lastRenderedContent && NoteCache.cachedRenderedHTML) {
                return NoteCache.cachedRenderedHTML;
            }
            
            // Strip YAML frontmatter from content before rendering
            let contentToRender = this.note.content;
            if (contentToRender.trim().startsWith('---')) {
                const lines = contentToRender.split('\n');
                if (lines[0].trim() === '---') {
                    // Find closing ---
                    let endIdx = -1;
                    for (let i = 1; i < lines.length; i++) {
                        if (lines[i].trim() === '---') {
                            endIdx = i;
                            break;
                        }
                    }
                    if (endIdx !== -1) {
                        // Remove frontmatter (including the closing ---) and any empty lines after it
                        contentToRender = lines.slice(endIdx + 1).join('\n').trim();
                    }
                }
            }
            
            // Convert Obsidian-style wikilinks: [[note]] or [[note|display text]]
            // Must be done before marked.parse() to avoid conflicts with markdown syntax
            // BUT we need to protect code blocks first to avoid converting [[text]] inside code
            const self = this; // Reference for closure
            
            // Step 1: Temporarily replace code blocks and inline code with placeholders
            const codeBlocks = [];
            // Protect fenced code blocks (```...```)
            contentToRender = contentToRender.replace(/```[\s\S]*?```/g, (match) => {
                codeBlocks.push(match);
                return `\x00CODEBLOCK${codeBlocks.length - 1}\x00`;
            });
            // Protect inline code (`...`)
            contentToRender = contentToRender.replace(/`[^`]+`/g, (match) => {
                codeBlocks.push(match);
                return `\x00CODEBLOCK${codeBlocks.length - 1}\x00`;
            });
            
            // Step 2: Convert media wikilinks FIRST: ![[file.png]] or ![[file.png|alt text]]
            // Must be before note wikilinks to prevent [[file.png]] from being matched first
            contentToRender = contentToRender.replace(
                /!\[\[([^\]|]+)(?:\|([^\]]+))?\]\]/g,
                (match, mediaName, altText) => {
                    const filename = mediaName.trim();
                    const alt = altText ? altText.trim() : filename.replace(/\.[^/.]+$/, '');
                    
                    // Resolve media path using O(1) lookup
                    const mediaPath = self.resolveMediaWikilink(filename);
                    
                    if (mediaPath) {
                        // URL-encode path segments for the API
                        const encodedPath = mediaPath.split('/').map(segment => {
                            try {
                                return encodeURIComponent(decodeURIComponent(segment));
                            } catch (e) {
                                return encodeURIComponent(segment);
                            }
                        }).join('/');
                        
                        const safeAlt = self.escapeHtml(alt).replace(/"/g, '&quot;');
                        const mediaSrc = `/api/media/${encodedPath}`;
                        const mediaType = self.getMediaType(filename);
                        
                        // Return appropriate HTML based on media type
                        switch (mediaType) {
                            case 'audio':
                                return `<div class="media-embed media-audio"><audio controls preload="none" src="${mediaSrc}" title="${safeAlt}"></audio><span class="media-caption">${safeAlt}</span></div>`;
                            case 'video':
                                return `<div class="media-embed media-video"><video controls preload="none" poster="" src="${mediaSrc}" title="${safeAlt}"></video></div>`;
                            case 'document':
                                // Local PDFs: show iframe preview
                                return `<div class="media-embed media-pdf"><iframe src="${mediaSrc}" title="${safeAlt}"></iframe></div>`;
                            default: // image
                                return `<img src="${mediaSrc}" alt="${safeAlt}" title="${safeAlt}">`;
                        }
                    }
                    
                    // Media not found - return broken indicator
                    const safeFilename = filename.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
                    const mediaType = self.getMediaType(filename);
                    const icon = mediaType === 'audio' ? '🎵' : mediaType === 'video' ? '🎬' : mediaType === 'document' ? '📄' : '🖼️';
                    return `<span class="wikilink-broken" title="${this.t('media.not_found')}">${icon} ${safeFilename}</span>`;
                }
            );
            
            // Step 2b: Convert note wikilinks: [[note]] or [[note|display text]]
            contentToRender = contentToRender.replace(
                /\[\[([^\]|]+)(?:\|([^\]]+))?\]\]/g,
                (match, target, displayText) => {
                    const linkTarget = target.trim();
                    const linkText = displayText ? displayText.trim() : linkTarget;
                    
                    // Fast O(1) check using pre-built lookup maps
                    // Handle section anchors: extract base note path
                    const hashIndex = linkTarget.indexOf('#');
                    const basePath = hashIndex !== -1 ? linkTarget.substring(0, hashIndex) : linkTarget;
                    const noteExists = basePath === '' || self.wikiLinkExists(basePath);
                    
                    // Escape special chars: href needs quote escaping, text needs HTML escaping
                    const safeHref = linkTarget.replace(/"/g, '%22');
                    const safeText = linkText.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
                    
                    // Return link with data attribute for styling broken links
                    const brokenClass = noteExists ? '' : ' class="wikilink-broken"';
                    return `<a href="${safeHref}"${brokenClass} data-wikilink="true">${safeText}</a>`;
                }
            );
            
            // Step 3: Restore code blocks
            contentToRender = contentToRender.replace(/\x00CODEBLOCK(\d+)\x00/g, (match, index) => {
                return codeBlocks[parseInt(index)];
            });

            // Create custom renderer to add IDs to headings and interactive checkboxes
            const renderer = new marked.Renderer();
            const slugCounts = {}; // Track duplicate slugs within this document

            // Custom heading renderer (same as before)
            renderer.heading = function(text, level, raw) {
                // Generate GitHub-style slug from heading text
                let slug = raw
                    .toLowerCase()
                    .replace(/[^\w\s-]/g, '') // Remove special chars
                    .replace(/\s+/g, '-')     // Spaces to dashes
                    .replace(/-+/g, '-');     // Multiple dashes to single

                // Handle duplicate slugs (same as extractOutline)
                if (slugCounts[slug] !== undefined) {
                    slugCounts[slug]++;
                    slug = `${slug}-${slugCounts[slug]}`;
                } else {
                    slugCounts[slug] = 0;
                }

                return `<h${level} id="${slug}">${text}</h${level}>\n`;
            };

            // Custom listitem renderer to make checkboxes interactive
            renderer.listitem = function(text, task, checked) {
                if (task) {
                    // Remove the default disabled checkbox from text and add interactive one
                    const cleanText = text.replace(/<input[^>]*type="checkbox"[^>]*>[ \t]*/, '');
                    const checkboxState = checked ? 'checked' : '';
                    return `<li data-task="true"><input type="checkbox" ${checkboxState} data-interactive-checkbox />${cleanText}</li>\n`;
                }
                return `<li>${text}</li>\n`;
            };

            // Custom tokenizer to disable Setext-style headings (lheading)
            // This prevents lines followed by "---" or "===" from being converted to headings
            // which causes issues when users write paths like "/api/auth/register" followed by "---"
            const customTokenizer = new marked.Tokenizer();
            customTokenizer.lheading = function() {
                return null; // Never match Setext-style headings
            };

            // Configure marked with syntax highlighting and custom renderer
            marked.setOptions({
                breaks: true,
                gfm: true,
                renderer: renderer,
                tokenizer: customTokenizer,
                highlight: function(code, lang) {
                    if (lang && hljs.getLanguage(lang)) {
                        try {
                            return hljs.highlight(code, { language: lang }).value;
                        } catch (err) {
                            console.error('Highlight error:', err);
                        }
                    }
                    return hljs.highlightAuto(code).value;
                }
            });
            
            // Parse markdown
            let html = marked.parse(contentToRender);

            // Post-process: Add target="_blank" to external links and title attributes to images
            // Parse as DOM to safely manipulate (DOMPurify sanitizes marked output to prevent XSS)
            const tempDiv = document.createElement('div');
            tempDiv.innerHTML = DOMPurify.sanitize(html, {
                ADD_TAGS: ['iframe'],
                ADD_ATTR: ['target', 'rel', 'controls', 'preload', 'poster', 'allowfullscreen']
            });
            
            // Find all links
            const links = tempDiv.querySelectorAll('a');
            links.forEach(link => {
                const href = link.getAttribute('href');
                if (href && typeof href === 'string') {
                    // Check if it's an external link
                    const isExternal = href.indexOf('http://') === 0 || 
                                      href.indexOf('https://') === 0 || 
                                      href.indexOf('//') === 0;
                    
                    if (isExternal) {
                        link.setAttribute('target', '_blank');
                        link.setAttribute('rel', 'noopener noreferrer');
                    }
                }
            });
            
            // Find all images and transform paths for display
            // Also convert non-image media (audio, video, PDF) to appropriate elements
            const images = tempDiv.querySelectorAll('img');
            images.forEach(img => {
                let src = img.getAttribute('src');
                if (src) {
                    const isExternal = src.startsWith('http://') || src.startsWith('https://') || src.startsWith('//');
                    const isLocal = !isExternal && !src.startsWith('data:');
                    
                    // Transform relative paths to /api/media/ for serving
                    if (isLocal && !src.startsWith('/api/media/')) {
                        // URL-encode path segments to handle spaces and special characters
                        const encodedPath = src.split('/').map(segment => {
                            try {
                                return encodeURIComponent(decodeURIComponent(segment));
                            } catch (e) {
                                return encodeURIComponent(segment);
                            }
                        }).join('/');
                        src = `/api/media/${encodedPath}`;
                        img.setAttribute('src', src);
                    }
                    
                    // Check if this is non-image media and convert to appropriate element
                    const mediaType = self.getMediaType(src);
                    const altText = img.getAttribute('alt') || src.split('/').pop().replace(/\.[^/.]+$/, '');
                    const safeAlt = self.escapeHtml(altText).replace(/"/g, '&quot;');
                    
                    // Only convert LOCAL media to embedded elements (security)
                    // External non-image media gets styled links instead
                    if (isLocal || src.startsWith('/api/media/')) {
                        if (mediaType === 'audio') {
                            const wrapper = document.createElement('div');
                            wrapper.className = 'media-embed media-audio';
                            const audioEl = document.createElement('audio');
                            audioEl.controls = true;
                            audioEl.preload = 'none';
                            audioEl.src = src;
                            audioEl.title = safeAlt || '';
                            const captionEl = document.createElement('span');
                            captionEl.className = 'media-caption';
                            captionEl.textContent = safeAlt || '';
                            wrapper.innerHTML = '';
                            wrapper.appendChild(audioEl);
                            wrapper.appendChild(captionEl);
                            img.replaceWith(wrapper);
                            return;
                        } else if (mediaType === 'video') {
                            const wrapper = document.createElement('div');
                            wrapper.className = 'media-embed media-video';
                            const videoEl = document.createElement('video');
                            videoEl.controls = true;
                            videoEl.preload = 'none';
                            videoEl.src = src;
                            videoEl.title = safeAlt || '';
                            wrapper.innerHTML = '';
                            wrapper.appendChild(videoEl);
                            img.replaceWith(wrapper);
                            return;
                        } else if (mediaType === 'document') {
                            // Local PDFs: show iframe preview
                            const wrapper = document.createElement('div');
                            wrapper.className = 'media-embed media-pdf';
                            const iframeEl = document.createElement('iframe');
                            iframeEl.src = src;
                            iframeEl.title = safeAlt || '';
                            wrapper.innerHTML = '';
                            wrapper.appendChild(iframeEl);
                            img.replaceWith(wrapper);
                            return;
                        }
                    } else if (isExternal && mediaType === 'document') {
                        // External PDFs: styled link (opens in new tab)
                        const link = document.createElement('a');
                        link.href = src;
                        link.target = '_blank';
                        link.rel = 'noopener noreferrer';
                        link.className = 'pdf-link';
                        link.title = `Open ${safeAlt}`;
                        link.innerHTML = `<span class="pdf-link-content">📄 ${safeAlt}</span><span class="pdf-link-note">Opens in new tab</span>`;
                        img.replaceWith(link);
                        return;
                    }
                    // External audio/video: leave as broken image for security
                }
                
                // For regular images, set title attribute
                const altText = img.getAttribute('alt');
                if (altText) {
                    img.setAttribute('title', altText);
                }
            });
            
            html = tempDiv.innerHTML;
            
            // Debounced MathJax rendering (avoid re-running on every keystroke)
            if (NoteCache.mathDebounceTimeout) clearTimeout(NoteCache.mathDebounceTimeout);
            NoteCache.mathDebounceTimeout = setTimeout(() => this.typesetMath(), 300);
            
            // Debounced Mermaid rendering
            if (NoteCache.mermaidDebounceTimeout) clearTimeout(NoteCache.mermaidDebounceTimeout);
            NoteCache.mermaidDebounceTimeout = setTimeout(() => this.renderMermaid(), 300);
            
            // Apply syntax highlighting and add copy buttons to code blocks
            setTimeout(() => {
                // Use cached reference if available, otherwise query
                const previewEl = NoteCache.domCache.previewContent || document.querySelector('.markdown-preview');
                if (previewEl) {
                    // Exclude code blocks that are rendered by other tools (e.g., Mermaid diagrams)
                    // Note: MathJax uses $$...$$ delimiters (not code blocks) so no exclusion needed
                    previewEl.querySelectorAll('pre code:not(.language-mermaid)').forEach((block) => {
                        // Apply syntax highlighting
                        if (!block.classList.contains('hljs')) {
                            hljs.highlightElement(block);
                        }
                        
                        // Add copy button if not already present
                        const pre = block.parentElement;
                        if (pre && !pre.querySelector('.copy-code-button')) {
                            this.addCopyButtonToCodeBlock(pre);
                        }
                    });
                    
                    // Enable video metadata loading (for first frame preview)
                    // Track by source URL to prevent duplicate requests on re-renders
                    if (!NoteCache.initializedVideoSources) NoteCache.initializedVideoSources = new Set();
                    previewEl.querySelectorAll('video[preload="none"]').forEach((video) => {
                        const src = video.getAttribute('src');
                        if (src && !NoteCache.initializedVideoSources.has(src)) {
                            NoteCache.initializedVideoSources.add(src);
                            video.preload = 'metadata';
                        }
                    });
                }
            }, 0);
            
            // Cache the result for performance
            NoteCache.lastRenderedContent = this.note.content;
            NoteCache.cachedRenderedHTML = html;
            
            return html;
        },
        
        // Refresh DOM element cache
        refreshDOMCache() {
            NoteCache.domCache.editor = document.querySelector('.editor-textarea');
            NoteCache.domCache.previewContent = document.querySelector('.markdown-preview');
            NoteCache.domCache.previewContainer = NoteCache.domCache.previewContent ? NoteCache.domCache.previewContent.parentElement : null;
            NoteCache.domCache.sidebar = document.querySelector('.flex-1.overflow-y-auto.custom-scrollbar');
            NoteCache.domCache.themeColorMeta = document.querySelector('meta[name="theme-color"]');
        },
        
        // Add copy button to code block
        addCopyButtonToCodeBlock(preElement) {
            // Extract language from code element class (e.g., "language-toml" -> "TOML")
            const codeElement = preElement.querySelector('code');
            let language = '';
            if (codeElement && codeElement.className) {
                const match = codeElement.className.match(/language-(\w+)/);
                if (match) {
                    const langMap = {
                        'js': 'JavaScript', 'ts': 'TypeScript', 'py': 'Python',
                        'rb': 'Ruby', 'cs': 'C#', 'cpp': 'C++', 'sh': 'Shell',
                        'bash': 'Bash', 'zsh': 'Zsh', 'yml': 'YAML', 'md': 'Markdown'
                    };
                    const rawLang = match[1].toLowerCase();
                    language = langMap[rawLang] || match[1].toUpperCase();
                }
            }
            
            // Create copy button with language label
            const button = document.createElement('button');
            button.className = 'copy-code-button';
            const displayText = language || this.t('common.copy_to_clipboard').split(' ')[0]; // Use first word as fallback
            button.innerHTML = `<span>${displayText}</span>`;
            button.dataset.originalText = displayText; // Store for restore after copy
            button.title = this.t('common.copy_to_clipboard');
            
            // Style the button
            button.style.position = 'absolute';
            button.style.top = '8px';
            button.style.right = '8px';
            button.style.padding = '4px 10px';
            button.style.backgroundColor = 'rgba(0, 0, 0, 0.6)';
            button.style.border = 'none';
            button.style.borderRadius = '4px';
            button.style.cursor = 'pointer';
            button.style.opacity = '0';
            button.style.transition = 'opacity 0.2s, background-color 0.2s';
            button.style.color = 'white';
            button.style.display = 'flex';
            button.style.alignItems = 'center';
            button.style.justifyContent = 'center';
            button.style.zIndex = '10';
            button.style.fontSize = '11px';
            button.style.fontWeight = '600';
            button.style.fontFamily = 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace';
            button.style.textTransform = 'uppercase';
            button.style.letterSpacing = '0.5px';
            
            // Style the pre element to be relative
            preElement.style.position = 'relative';
            
            // Show button on hover
            preElement.addEventListener('mouseenter', () => {
                button.style.opacity = '1';
            });
            
            preElement.addEventListener('mouseleave', () => {
                button.style.opacity = '0';
            });
            
            // Copy to clipboard on click
            button.addEventListener('click', async (e) => {
                e.preventDefault();
                e.stopPropagation();
                
                const codeElement = preElement.querySelector('code');
                if (!codeElement) return;
                
                const code = codeElement.textContent;
                
                const originalText = button.dataset.originalText;
                const copiedText = this.t('common.copied');
                const copyTitle = this.t('common.copy_to_clipboard');
                
                try {
                    await navigator.clipboard.writeText(code);
                    
                    // Visual feedback - show localized "Copied!"
                    button.innerHTML = `<span>${copiedText}</span>`;
                    button.style.backgroundColor = 'rgba(34, 197, 94, 0.8)';
                    button.title = copiedText;
                    
                    // Reset after 2 seconds
                    setTimeout(() => {
                        button.innerHTML = `<span>${originalText}</span>`;
                        button.style.backgroundColor = 'rgba(0, 0, 0, 0.6)';
                        button.title = copyTitle;
                    }, 2000);
                } catch (err) {
                    console.error('Failed to copy code:', err);
                    
                    // Fallback for older browsers
                    const textArea = document.createElement('textarea');
                    textArea.value = code;
                    textArea.style.position = 'fixed';
                    textArea.style.left = '-999999px';
                    document.body.appendChild(textArea);
                    textArea.select();
                    
                    try {
                        document.execCommand('copy');
                        button.innerHTML = `<span>${copiedText}</span>`;
                        button.style.backgroundColor = 'rgba(34, 197, 94, 0.8)';
                        setTimeout(() => {
                            button.innerHTML = `<span>${originalText}</span>`;
                            button.style.backgroundColor = 'rgba(0, 0, 0, 0.6)';
                        }, 2000);
                    } catch (fallbackErr) {
                        console.error('Fallback copy failed:', fallbackErr);
                    }
                    
                    document.body.removeChild(textArea);
                }
            });
            
            // Add button to pre element
            preElement.appendChild(button);
        },
        
        // Setup scroll synchronization
        setupScrollSync() {
            // Use cached references (refresh if not available)
            if (!NoteCache.domCache.editor || !NoteCache.domCache.previewContainer) {
                this.refreshDOMCache();
            }
            
            const editor = NoteCache.domCache.editor;
            const preview = NoteCache.domCache.previewContainer;
            
            if (!editor || !preview) {
                // If elements don't exist yet, retry with limit
                if (!this._setupScrollSyncRetries) this._setupScrollSyncRetries = 0;
                if (this._setupScrollSyncRetries < CONFIG.SCROLL_SYNC_MAX_RETRIES) {
                    this._setupScrollSyncRetries++;
                    setTimeout(() => this.setupScrollSync(), CONFIG.SCROLL_SYNC_RETRY_INTERVAL);
                } else {
                    console.warn(`setupScrollSync: Failed to find editor/preview elements after ${CONFIG.SCROLL_SYNC_MAX_RETRIES} retries`);
                }
                return;
            }
            
            // Reset retry counter on success
            this._setupScrollSyncRetries = 0;
            
            // Remove old listeners if they exist
            if (this._editorScrollHandler) {
                editor.removeEventListener('scroll', this._editorScrollHandler);
            }
            if (this._previewScrollHandler) {
                preview.removeEventListener('scroll', this._previewScrollHandler);
            }
            
            // Create new scroll handlers with heading-anchored sync
            // Key improvement: uses heading positions as anchor points for precise alignment
            // between editor text lines and rendered preview elements. Falls back to
            // percentage-based sync when no heading is near the viewport.
            this._editorScrollHandler = () => {
                if (this.ui.isScrolling) return;
                
                const scrollableHeight = editor.scrollHeight - editor.clientHeight;
                if (scrollableHeight <= 0) return;
                
                const previewScrollableHeight = preview.scrollHeight - preview.clientHeight;
                if (previewScrollableHeight <= 0) return;
                
                let targetScrollTop = null;
                
                // Try heading-anchored sync first
                const heading = this.findNearestHeadingInEditor(editor);
                if (heading) {
                    const headingEl = document.getElementById(heading.slug);
                    if (headingEl && preview.contains(headingEl)) {
                        const containerRect = preview.getBoundingClientRect();
                        const headingRect = headingEl.getBoundingClientRect();
                        targetScrollTop = preview.scrollTop + (headingRect.top - containerRect.top);
                    }
                }
                
                // Fallback to percentage-based sync
                if (targetScrollTop === null) {
                    const scrollPercentage = editor.scrollTop / scrollableHeight;
                    targetScrollTop = scrollPercentage * previewScrollableHeight;
                }
                
                this.ui.isScrolling = true;
                preview.scrollTop = Math.max(0, Math.min(targetScrollTop, preview.scrollHeight - preview.clientHeight));
                setTimeout(() => {
                    this.ui.isScrolling = false;
                }, CONFIG.SCROLL_SYNC_DELAY);
            };
            
            this._previewScrollHandler = () => {
                if (this.ui.isScrolling) return;
                
                const scrollableHeight = preview.scrollHeight - preview.clientHeight;
                if (scrollableHeight <= 0) return;
                
                const editorScrollableHeight = editor.scrollHeight - editor.clientHeight;
                if (editorScrollableHeight <= 0) return;
                
                let targetScrollTop = null;
                
                // Try heading-anchored sync first
                const heading = this.findNearestHeadingInPreview(preview);
                if (heading && heading.line) {
                    const lineHeight = parseFloat(getComputedStyle(editor).lineHeight) || 24;
                    const paddingTop = parseFloat(getComputedStyle(editor).paddingTop) || 24;
                    targetScrollTop = (heading.line - 1) * lineHeight + paddingTop - editor.clientHeight / 3;
                }
                
                // Fallback to percentage-based sync
                if (targetScrollTop === null) {
                    const scrollPercentage = preview.scrollTop / scrollableHeight;
                    targetScrollTop = scrollPercentage * editorScrollableHeight;
                }
                
                this.ui.isScrolling = true;
                editor.scrollTop = Math.max(0, Math.min(targetScrollTop, editor.scrollHeight - editor.clientHeight));
                setTimeout(() => {
                    this.ui.isScrolling = false;
                }, CONFIG.SCROLL_SYNC_DELAY);
            };
            
            // Attach new listeners
            editor.addEventListener('scroll', this._editorScrollHandler);
            preview.addEventListener('scroll', this._previewScrollHandler);
        },
        
        // Calculate note statistics (client-side)
        calculateStats() {
            if (!this.note.content) {
                this.stats.data = null;
                return;
            }
            
            const content = this.note.content;
            
            // Word count
            const words = (content.match(/\S+/g) || []).length;
            
            // Character count
            const chars = content.replace(/\s/g, '').length;
            const totalChars = content.length;
            
            // Reading time (200 words per minute)
            const readingTime = Math.max(1, Math.round(words / 200));
            
            // Line count
            const lines = content.split('\n').length;
            
            // Paragraph count
            const paragraphs = content.split('\n\n').filter(p => p.trim()).length;
            
            // Sentences: punctuation [.!?]+ followed by space or end-of-string
            const sentences = (content.match(/[.!?]+(?:\s|$)/g) || []).length;
            
            // List items: lines starting with -, *, + or a number (e.g. 1., 10.), excluding tasks [-]
            const listItems = (content.match(/^\s*(?:[-*+]|\d+\.)\s+(?!\[)/gm) || []).length;
            
            // Tables: markdown table separator rows (| --- | --- |)
            const tables = (content.match(/^\s*\|(?:\s*:?-+:?\s*\|){1,}\s*$/gm) || []).length;
            
            // Link count (standard markdown links)
            const markdownLinkMatches = content.match(/\[([^\]]+)\]\(([^\)]+)\)/g) || [];
            const markdownLinks = markdownLinkMatches.length;
            const markdownInternalLinks = markdownLinkMatches.filter(l => l.includes('.md')).length;
            
            // Wikilink count ([[note]] or [[note|display text]] format)
            const wikilinks = (content.match(/\[\[([^\]|]+)(?:\|[^\]]+)?\]\]/g) || []).length;
            
            // Total links (markdown + wikilinks)
            const links = markdownLinks + wikilinks;
            const internalLinks = markdownInternalLinks + wikilinks; // All wikilinks are internal
            
            // Code blocks
            const codeBlocks = (content.match(/```[\s\S]*?```/g) || []).length;
            const inlineCode = (content.match(/`[^`]+`/g) || []).length;
            
            // Headings
            const h1 = (content.match(/^# /gm) || []).length;
            const h2 = (content.match(/^## /gm) || []).length;
            const h3 = (content.match(/^### /gm) || []).length;
            
            // Tasks
            const totalTasks = (content.match(/- \[[ x]\]/gi) || []).length;
            const completedTasks = (content.match(/- \[x\]/gi) || []).length;
            const pendingTasks = totalTasks - completedTasks;
            const completionRate = totalTasks > 0 ? Math.round((completedTasks / totalTasks) * 100) : 0;
            
            // Images
            const images = (content.match(/!\[([^\]]*)\]\(([^\)]+)\)/g) || []).length;
            
            // Blockquotes
            const blockquotes = (content.match(/^> /gm) || []).length;
            
            this.stats.data = {
                words,
                sentences,
                characters: chars,
                total_characters: totalChars,
                reading_time_minutes: readingTime,
                lines,
                paragraphs,
                list_items: listItems,
                tables,
                links,
                internal_links: internalLinks,
                external_links: links - internalLinks,
                wikilinks,
                code_blocks: codeBlocks,
                inline_code: inlineCode,
                headings: {
                    h1,
                    h2,
                    h3,
                    total: h1 + h2 + h3
                },
                tasks: {
                    total: totalTasks,
                    completed: completedTasks,
                    pending: pendingTasks,
                    completion_rate: completionRate
                },
                images,
                blockquotes
            };
        },
        
        // Parse YAML frontmatter metadata from note content
        parseMetadata() {
            if (!this.note.content) {
                this.note.metadata = null;
                NoteCache.lastFrontmatter = null;
                return;
            }
            
            const content = this.note.content;
            
            // Check if content starts with frontmatter
            if (!content.trim().startsWith('---')) {
                this.note.metadata = null;
                NoteCache.lastFrontmatter = null;
                return;
            }
            
            try {
                const lines = content.split('\n');
                if (lines[0].trim() !== '---') {
                    this.note.metadata = null;
                    NoteCache.lastFrontmatter = null;
                    return;
                }
                
                // Find closing ---
                let endIdx = -1;
                for (let i = 1; i < lines.length; i++) {
                    if (lines[i].trim() === '---') {
                        endIdx = i;
                        break;
                    }
                }
                
                if (endIdx === -1) {
                    this.note.metadata = null;
                    NoteCache.lastFrontmatter = null;
                    return;
                }
                
                // Performance optimization: skip parsing if frontmatter unchanged
                const frontmatterRaw = lines.slice(0, endIdx + 1).join('\n');
                if (frontmatterRaw === NoteCache.lastFrontmatter) {
                    return; // No change, keep existing metadata
                }
                NoteCache.lastFrontmatter = frontmatterRaw;
                
                const frontmatterLines = lines.slice(1, endIdx);
                const metadata = {};
                let currentKey = null;
                let currentValue = [];
                
                for (const line of frontmatterLines) {
                    // Check for new key: value pair (supports keys with hyphens/underscores)
                    const keyMatch = line.match(/^([a-zA-Z_][\w-]*):\s*(.*)$/);
                    
                    if (keyMatch) {
                        // Save previous key if exists
                        if (currentKey) {
                            metadata[currentKey] = this.parseYamlValue(currentValue.join('\n'));
                        }
                        
                        currentKey = keyMatch[1];
                        const value = keyMatch[2].trim();
                        currentValue = [value];
                    } else if (line.match(/^\s+-\s+/) && currentKey) {
                        // List item continuation (e.g., "  - item")
                        currentValue.push(line);
                    } else if (line.startsWith('  ') && currentKey) {
                        // Indented content (multiline value)
                        currentValue.push(line);
                    }
                }
                
                // Save last key
                if (currentKey) {
                    metadata[currentKey] = this.parseYamlValue(currentValue.join('\n'));
                }
                
                this.note.metadata = Object.keys(metadata).length > 0 ? metadata : null;
                
            } catch (error) {
                console.error('Failed to parse frontmatter:', error);
                this.note.metadata = null;
                NoteCache.lastFrontmatter = null;
            }
        },
        
        async loadAttachments() {
            if (!this.note.current) {
                this.media.attachments = [];
                return;
            }
            
        this.media.attachmentsLoading = true;
            
        try {
            const encodedPath = this.note.current.split('/').map(segment => {
                try {
                    return encodeURIComponent(decodeURIComponent(segment));
                } catch (e) {
                    return encodeURIComponent(segment);
                }
            }).join('/');
            const response = await fetch(`/api/notes/attachments/${encodedPath}`);
                
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                
                const data = await response.json();
                this.media.attachments = data.attachments || [];
            } catch (error) {
                console.error('Failed to load attachments:', error);
                this.media.attachments = [];
            } finally {
                this.media.attachmentsLoading = false;
            }
        },
        
        // Parse a YAML value (handles arrays, strings, numbers, booleans)
        parseYamlValue(value) {
            if (!value || value.trim() === '') return null;
            
            value = value.trim();
            
            // Check for inline array: [item1, item2]
            if (value.startsWith('[') && value.endsWith(']')) {
                const inner = value.slice(1, -1);
                return inner.split(',').map(s => s.trim().replace(/^["']|["']$/g, '')).filter(s => s);
            }
            
            // Check for YAML list format (multiple lines starting with -)
            if (value.includes('\n  -') || value.startsWith('  -')) {
                const items = [];
                const lines = value.split('\n');
                for (const line of lines) {
                    const match = line.match(/^\s*-\s*(.+)$/);
                    if (match) {
                        items.push(match[1].trim().replace(/^["']|["']$/g, ''));
                    }
                }
                return items.length > 0 ? items : value;
            }
            
            // Check for boolean
            if (value.toLowerCase() === 'true') return true;
            if (value.toLowerCase() === 'false') return false;
            
            // Check for number
            if (/^-?\d+(\.\d+)?$/.test(value)) {
                return parseFloat(value);
            }
            
            // Return as string (remove quotes if present)
            return value.replace(/^["']|["']$/g, '');
        },
        
        // Check if a string is a URL
        isUrl(str) {
            if (typeof str !== 'string') return false;
            return /^https?:\/\/\S+$/i.test(str.trim());
        },
        
        // Escape HTML to prevent XSS
        escapeHtml(str) {
            const div = document.createElement('div');
            div.textContent = str;
            return div.innerHTML;
        },
        
        // Format metadata value for display
        formatMetadataValue(key, value) {
            if (value === null || value === undefined) return '';
            
            // Arrays are handled separately in the template
            if (Array.isArray(value)) return value;
            
            // Format dates nicely
            if (key === 'date' || key === 'created' || key === 'modified' || key === 'updated') {
                let date;
                // Parse date-only strings (YYYY-MM-DD) as local dates to avoid timezone issues
                if (typeof value === 'string' && /^\d{4}-\d{2}-\d{2}$/.test(value)) {
                    const [year, month, day] = value.split('-').map(Number);
                    date = new Date(year, month - 1, day);  // month is 0-indexed
                } else {
                    date = new Date(value);
                }
                if (!isNaN(date.getTime())) {
                    return date.toLocaleDateString(this.i18n.locale, { 
                        year: 'numeric', 
                        month: 'short', 
                        day: 'numeric' 
                    });
                }
            }
            
            // Booleans
            if (typeof value === 'boolean') {
                return value ? this.t('common.yes') : this.t('common.no');
            }
            
            return String(value);
        },
        
        // Format metadata value as HTML (for URL support)
        formatMetadataValueHtml(key, value) {
            const formatted = this.formatMetadataValue(key, value);
            
            // Check if it's a URL
            if (this.isUrl(formatted)) {
                const escaped = this.escapeHtml(formatted);
                // Truncate long URLs for display
                const displayUrl = formatted.length > 40 
                    ? formatted.substring(0, 37) + '...' 
                    : formatted;
                return `<a href="${escaped}" target="_blank" rel="noopener noreferrer" class="metadata-link">${this.escapeHtml(displayUrl)}</a>`;
            }
            
            return this.escapeHtml(formatted);
        },
        
        // Get priority metadata fields (shown in collapsed view)
        getPriorityMetadataFields() {
            if (!this.note.metadata) return [];
            
            // Fields to show in collapsed view, in order of priority
            const priority = ['date', 'created', 'author', 'status', 'priority', 'type', 'category'];
            const fields = [];
            
            for (const key of priority) {
                if (this.note.metadata[key] !== undefined && !Array.isArray(this.note.metadata[key])) {
                    const formatted = this.formatMetadataValue(key, this.note.metadata[key]);
                    const isUrl = this.isUrl(formatted);
                    fields.push({ 
                        key, 
                        value: formatted,
                        valueHtml: isUrl ? this.formatMetadataValueHtml(key, this.note.metadata[key]) : this.escapeHtml(formatted),
                        isUrl
                    });
                }
            }
            
            return fields.slice(0, 3); // Show max 3 fields in collapsed view
        },
        
        // Get all metadata fields except tags (for expanded view)
        getAllMetadataFields() {
            if (!this.note.metadata) return [];
            
            return Object.entries(this.note.metadata)
                .filter(([key]) => key !== 'tags') // Tags shown separately
                .map(([key, value]) => {
                    const isArray = Array.isArray(value);
                    const formatted = this.formatMetadataValue(key, value);
                    const isUrl = !isArray && this.isUrl(formatted);
                    return {
                        key,
                        value: formatted,
                        valueHtml: isUrl ? this.formatMetadataValueHtml(key, value) : this.escapeHtml(formatted),
                        isArray,
                        isUrl
                    };
                });
        },
        
        // Check if note has any displayable metadata
        getHasMetadata() {
            const has = this.note.metadata && Object.keys(this.note.metadata).length > 0;
            return has;
        },
        
        // Get tags from metadata
        getMetadataTags() {
            if (!this.note.metadata || !this.note.metadata.tags) return [];
            return Array.isArray(this.note.metadata.tags) ? this.note.metadata.tags : [this.note.metadata.tags];
        },
        
        // Save sidebar width to localStorage
        saveSidebarWidth() {
            localStorage.setItem('sidebarWidth', this.ui.sidebarWidth.toString());
        },
        
        // Save view mode to localStorage
        saveViewMode() {
            try {
                localStorage.setItem('viewMode', this.ui.viewMode);
            } catch (error) {
                console.error('Error saving view mode:', error);
            }
        },
        
        saveTagsExpanded() {
            try {
                localStorage.setItem('tagsExpanded', this.tags.expanded.toString());
            } catch (error) {
                console.error('Error saving tags expanded state:', error);
            }
        },
        
        saveExpandedFolders() {
            try {
                localStorage.setItem('expandedFolders', JSON.stringify([...this.folders.expanded]));
            } catch (error) {
                console.error('Error saving expanded folders:', error);
            }
        },
        
        // Start resizing sidebar
        startResize(event) {
            this.ui.isResizing = true;
            event.preventDefault();
            
            const resize = (e) => {
                if (!this.ui.isResizing) return;
                
                // Calculate new width based on mouse position
                const newWidth = e.clientX;
                
                // Clamp between min and max
                if (newWidth >= 200 && newWidth <= 600) {
                    this.ui.sidebarWidth = newWidth;
                }
            };
            
            const stopResize = () => {
                if (this.ui.isResizing) {
                    this.ui.isResizing = false;
                    this.saveSidebarWidth();
                    document.removeEventListener('mousemove', resize);
                    document.removeEventListener('mouseup', stopResize);
                }
            };
            
            document.addEventListener('mousemove', resize);
            document.addEventListener('mouseup', stopResize);
        },
        
        // Start resizing split panes (editor/preview)
        startSplitResize(event) {
            this.ui.isResizingSplit = true;
            event.preventDefault();
            
            const container = event.target.parentElement;
            const HANDLE_WIDTH = 6; // Match CSS width of .split-resize-handle
            const SNAP_THRESHOLD = 3; // Snap to 50% when within 3%
            
            const getClientX = (e) => {
                if (e.touches && e.touches.length > 0) return e.touches[0].clientX;
                return e.clientX;
            };
            
            const resize = (e) => {
                if (!this.ui.isResizingSplit) return;
                e.preventDefault();
                
                const containerRect = container.getBoundingClientRect();
                const clientX = getClientX(e);
                const mouseX = clientX - containerRect.left - (HANDLE_WIDTH / 2);
                let percentage = (mouseX / containerRect.width) * 100;
                
                // Snap to 50% midpoint when close
                if (Math.abs(percentage - 50) < SNAP_THRESHOLD) {
                    percentage = 50;
                }
                
                // Clamp between 20% and 80%
                if (percentage >= 20 && percentage <= 80) {
                    this.ui.editorWidth = percentage;
                }
            };
            
            const stopResize = () => {
                if (this.ui.isResizingSplit) {
                    this.ui.isResizingSplit = false;
                    this.saveEditorWidth();
                    document.removeEventListener('mousemove', resize);
                    document.removeEventListener('mouseup', stopResize);
                    document.removeEventListener('touchmove', resize);
                    document.removeEventListener('touchend', stopResize);
                }
            };
            
            document.addEventListener('mousemove', resize);
            document.addEventListener('mouseup', stopResize);
            document.addEventListener('touchmove', resize, { passive: false });
            document.addEventListener('touchend', stopResize);
        },
        
        // Setup mobile view mode handler (auto-switch from split to edit on mobile)
        setupMobileViewMode() {
            const MOBILE_BREAKPOINT = 768; // Match CSS breakpoint
            let previousWidth = window.innerWidth;
            
            const handleResize = () => {
                const currentWidth = window.innerWidth;
                const wasMobile = previousWidth <= MOBILE_BREAKPOINT;
                const isMobile = currentWidth <= MOBILE_BREAKPOINT;
                
                // If switching from desktop to mobile and in split mode
                if (!wasMobile && isMobile && this.ui.viewMode === 'split') {
                    this.ui.viewMode = 'edit';
                }
                
                previousWidth = currentWidth;
            };
            
            // Listen for window resize
            window.addEventListener('resize', handleResize);
            
            // Check initial state
            if (window.innerWidth <= MOBILE_BREAKPOINT && this.ui.viewMode === 'split') {
                this.ui.viewMode = 'edit';
            }
            
            // Setup swipe gestures for mobile
            this.setupSwipeGestures();
            this.setupTouchEnhancements();
        },
        
        // Setup swipe gestures for mobile sidebar
        setupSwipeGestures() {
            const SWIPE_THRESHOLD = 50; // Minimum distance for swipe
            const EDGE_THRESHOLD = 40;  // Distance from edge to start swipe (increased for better UX)
            
            let touchStartX = 0;
            let touchStartY = 0;
            let isSwiping = false;
            
            document.addEventListener('touchstart', (e) => {
                // Only handle on mobile
                if (window.innerWidth > 768) return;
                
                touchStartX = e.touches[0].clientX;
                touchStartY = e.touches[0].clientY;
                isSwiping = true;
            }, { passive: true });
            
            document.addEventListener('touchmove', (e) => {
                if (!isSwiping || window.innerWidth > 768) return;
                
                const touchX = e.touches[0].clientX;
                const touchY = e.touches[0].clientY;
                const deltaX = touchX - touchStartX;
                const deltaY = touchY - touchStartY;
                
                // Ignore vertical swipes
                if (Math.abs(deltaY) > Math.abs(deltaX)) {
                    isSwiping = false;
                    return;
                }
            }, { passive: true });
            
            document.addEventListener('touchend', (e) => {
                if (!isSwiping || window.innerWidth > 768) return;
                
                const touchEndX = e.changedTouches[0].clientX;
                const deltaX = touchEndX - touchStartX;
                
                // Swipe right from left edge to open sidebar
                if (touchStartX < EDGE_THRESHOLD && deltaX > SWIPE_THRESHOLD && !this.ui.mobileSidebarOpen) {
                    this.ui.mobileSidebarOpen = true;
                    // Haptic feedback if available
                    if (navigator.vibrate) navigator.vibrate(10);
                }
                
                // Swipe left to close sidebar
                if (this.ui.mobileSidebarOpen && deltaX < -SWIPE_THRESHOLD) {
                    this.ui.mobileSidebarOpen = false;
                    if (navigator.vibrate) navigator.vibrate(10);
                }
                
                isSwiping = false;
            }, { passive: true });
        },
        
        // Additional touch enhancements for mobile
        setupTouchEnhancements() {
            // Double tap to toggle edit/preview mode on mobile
            let lastTapTime = 0;
            const DOUBLE_TAP_DELAY = 300;
            
            document.addEventListener('touchend', (e) => {
                if (window.innerWidth > 768) return;
                
                // Only in editor/preview area
                const target = e.target;
                const isEditorArea = target.closest('.editor-textarea, .markdown-preview, .editor-wrapper');
                if (!isEditorArea) return;
                
                const currentTime = new Date().getTime();
                const tapLength = currentTime - lastTapTime;
                
                if (tapLength < DOUBLE_TAP_DELAY && tapLength > 0) {
                    // Double tap detected - toggle view mode
                    if (this.note.current) {
                        this.ui.viewMode = this.ui.viewMode === 'edit' ? 'preview' : 'edit';
                        if (navigator.vibrate) navigator.vibrate(10);
                    }
                    e.preventDefault();
                }
                
                lastTapTime = currentTime;
            }, { passive: false });
            
            // Add touch feedback to interactive elements
            document.addEventListener('touchstart', (e) => {
                if (window.innerWidth > 768) return;
                
                const interactive = e.target.closest('button, .note-item, .folder-item, .hover-accent, .tag-chip, .mobile-bottom-tab');
                if (interactive) {
                    interactive.classList.add('touch-active');
                }
            }, { passive: true });
            
            document.addEventListener('touchend', (e) => {
                if (window.innerWidth > 768) return;
                
                const activeElements = document.querySelectorAll('.touch-active');
                activeElements.forEach(el => el.classList.remove('touch-active'));
            }, { passive: true });
            
            // Prevent zoom on double tap for form inputs (iOS)
            document.addEventListener('touchend', (e) => {
                const now = Date.now();
                if (now - lastTapTime <= 300) {
                    const isInput = e.target.matches('input, textarea, select');
                    if (isInput) {
                        e.preventDefault();
                    }
                }
            }, { passive: false });
        },
        
        // Save editor width to localStorage
        saveEditorWidth() {
            localStorage.setItem('editorWidth', this.ui.editorWidth.toString());
        },
        
        // Save scroll position for current note
        saveScrollPosition() {
            if (!this.note.current) return;
            
            // Refresh DOM cache if needed
            if (!NoteCache.domCache.editor || !NoteCache.domCache.previewContainer) {
                this.refreshDOMCache();
            }
            
            const editorScroll = NoteCache.domCache.editor ? NoteCache.domCache.editor.scrollTop : 0;
            const previewScroll = NoteCache.domCache.previewContainer ? NoteCache.domCache.previewContainer.scrollTop : 0;
            
            NoteCache.setScrollPosition(this.note.current, editorScroll, previewScroll);
        },
        
        // Restore scroll position for current note
        restoreScrollPosition() {
            if (!this.note.current) return;
            
            // Get saved position
            const position = NoteCache.getScrollPosition(this.note.current);
            
            // Refresh DOM cache if needed
            if (!NoteCache.domCache.editor || !NoteCache.domCache.previewContainer) {
                this.refreshDOMCache();
            }
            
            // Disable scroll sync temporarily
            this.ui.isScrolling = true;
            
            // Restore scroll positions based on view mode
            if (this.ui.viewMode === 'edit' || this.ui.viewMode === 'split') {
                if (NoteCache.domCache.editor) {
                    NoteCache.domCache.editor.scrollTop = position.editor;
                }
            }
            
            if (this.ui.viewMode === 'preview' || this.ui.viewMode === 'split') {
                if (NoteCache.domCache.previewContainer) {
                    NoteCache.domCache.previewContainer.scrollTop = position.preview;
                }
            }
            
            // Re-enable scroll sync after a short delay
            setTimeout(() => {
                this.ui.isScrolling = false;
            }, CONFIG.SCROLL_SYNC_DELAY);
        },
        
        // Scroll to top of editor and preview
        scrollToTop() {
            // Disable scroll sync temporarily to prevent interference
            this.ui.isScrolling = true;
            
            // Use cached references (refresh if not available)
            if (!NoteCache.domCache.editor || !NoteCache.domCache.previewContainer) {
                this.refreshDOMCache();
            }
            
            // Only scroll the visible panes based on viewMode
            if (this.ui.viewMode === 'edit' || this.ui.viewMode === 'split') {
                if (NoteCache.domCache.editor) {
                    NoteCache.domCache.editor.scrollTop = 0;
                }
            }
            
            if (this.ui.viewMode === 'preview' || this.ui.viewMode === 'split') {
                // Scroll the preview container (parent of .markdown-preview)
                if (NoteCache.domCache.previewContainer) {
                    NoteCache.domCache.previewContainer.scrollTop = 0;
                }
            }
            
            // Re-enable scroll sync after a short delay
            setTimeout(() => {
                this.ui.isScrolling = false;
            }, CONFIG.SCROLL_SYNC_DELAY);
        },
        
        // Export current note as HTML
        async exportToHTML() {
            if (!this.note.current || !this.note.content) {
                this.showAlert(this.t('notes.no_content'));
                return;
            }
            
            try {
                // Get the note name without extension
                const noteName = this.note.name || 'note';
                
                // Get current rendered HTML (this already has markdown converted and will have LaTeX delimiters)
                let renderedHTML = this.renderedMarkdown;
                
                // Convert non-image media (audio, video, PDF) to placeholders first
                // These shouldn't be embedded as base64 (too large)
                // Use CSS variables with fallbacks for theme-aware styling (matches backend export.py)
                const mediaPlaceholder = (type, name) => {
                    const icons = { audio: '🎵', video: '🎬', document: '📄' };
                    const labels = { audio: 'Audio file', video: 'Video file', document: 'PDF document' };
                    const icon = icons[type] || '📎';
                    const label = labels[type] || 'Media file';
                    return `<div style="margin:1.5rem 0;padding:1.5rem;background:linear-gradient(135deg,var(--bg-tertiary,#f8f9fa) 0%,var(--bg-secondary,#e9ecef) 100%);border:1px solid var(--border-primary,#dee2e6);border-radius:0.5rem;display:flex;align-items:center;gap:1rem;">
<span style="font-size:2rem;">${icon}</span>
<div>
<div style="font-weight:600;color:var(--text-primary,#212529);">${name}</div>
<div style="font-size:0.875rem;color:var(--text-secondary,#6c757d);">${label} — not available in exported view</div>
</div>
</div>`;
                };
                
                // Replace audio embeds with placeholders
                renderedHTML = renderedHTML.replace(
                    /<div class="media-embed media-audio">.*?<audio[^>]*src="[^"]*\/([^"\/]+)"[^>]*>.*?<\/div>/gs,
                    (match, filename) => mediaPlaceholder('audio', decodeURIComponent(filename).replace(/\.[^.]+$/, ''))
                );
                
                // Replace video embeds with placeholders
                renderedHTML = renderedHTML.replace(
                    /<div class="media-embed media-video">.*?<video[^>]*src="[^"]*\/([^"\/]+)"[^>]*>.*?<\/div>/gs,
                    (match, filename) => mediaPlaceholder('video', decodeURIComponent(filename).replace(/\.[^.]+$/, ''))
                );
                
                // Replace PDF embeds with placeholders
                renderedHTML = renderedHTML.replace(
                    /<div class="media-embed media-pdf">.*?<iframe[^>]*src="[^"]*\/([^"\/]+)"[^>]*>.*?<\/div>/gs,
                    (match, filename) => mediaPlaceholder('document', decodeURIComponent(filename).replace(/\.[^.]+$/, ''))
                );
                
                // Embed local images as base64 for fully self-contained HTML
                // Handle both /api/media/ and legacy /api/images/ paths
                const imgRegex = /src="\/api\/(?:media|images)\/([^"]+)"/g;
                const imgMatches = [...renderedHTML.matchAll(imgRegex)];
                
                for (const match of imgMatches) {
                    const encodedPath = match[1];
                    // Skip non-image files (already handled above)
                    const ext = encodedPath.split('.').pop().toLowerCase();
                    if (!['jpg', 'jpeg', 'png', 'gif', 'webp'].includes(ext)) {
                        continue;
                    }
                    
                    try {
                        // Fetch the image
                        const imgResponse = await fetch(`/api/media/${encodedPath}`);
                        if (imgResponse.ok) {
                            const blob = await imgResponse.blob();
                            // Convert to base64 data URL
                            const base64 = await new Promise((resolve) => {
                                const reader = new FileReader();
                                reader.onloadend = () => resolve(reader.result);
                                reader.readAsDataURL(blob);
                            });
                            // Replace the src with base64 data URL
                            renderedHTML = renderedHTML.replace(match[0], `src="${base64}"`);
                        }
                    } catch (e) {
                        console.warn(`Failed to embed image: ${encodedPath}`, e);
                        // Fall back to relative path
                        const decodedPath = decodeURIComponent(encodedPath);
                        renderedHTML = renderedHTML.replace(match[0], `src="${decodedPath}"`);
                    }
                }
                
                // Get current theme CSS
                const currentTheme = this.theme.current || 'light';
                const themeResponse = await fetch(`/api/themes/${currentTheme}`);
                const themeText = await themeResponse.text();
                
                // Check if response is JSON or plain CSS
                let themeCss;
                try {
                    const themeJson = JSON.parse(themeText);
                    // If it's JSON, extract the css field
                    themeCss = themeJson.css || themeText;
                } catch (e) {
                    // If it's not JSON, use it as-is
                    themeCss = themeText;
                }
                
                // Theme CSS uses :root[data-theme="..."] selector, but we need plain :root for export
                // Strip the data-theme attribute selector so variables apply globally
                themeCss = themeCss.replace(/:root\[data-theme="[^"]+"\]/g, ':root');
                
                // Get highlight.js theme URL from current page
                const highlightLinkElement = document.getElementById('highlight-theme');
                if (!highlightLinkElement || !highlightLinkElement.href) {
                    console.warn('Could not detect highlight.js theme, export may not match preview exactly');
                }
                const highlightTheme = highlightLinkElement ? highlightLinkElement.href : '';
                
                // Extract all markdown preview styles from current page
                let markdownStyles = '';
                const styleSheets = Array.from(document.styleSheets);
                
                for (const sheet of styleSheets) {
                    try {
                        // Skip external stylesheets (CDN resources) to avoid CORS errors
                        // We link them directly in the exported HTML anyway
                        if (sheet.href && (sheet.href.startsWith('http://') || sheet.href.startsWith('https://'))) {
                            const currentOrigin = window.location.origin;
                            const sheetURL = new URL(sheet.href);
                            if (sheetURL.origin !== currentOrigin) {
                                // Skip cross-origin stylesheets (they're linked directly in export)
                                continue;
                            }
                        }
                        
                        const rules = Array.from(sheet.cssRules || []);
                        for (const rule of rules) {
                            const cssText = rule.cssText;
                            // Include rules that target markdown-preview, mjx-container, or mermaid-rendered
                            if (cssText.includes('.markdown-preview') || 
                                cssText.includes('mjx-container') ||
                                cssText.includes('.MathJax') ||
                                cssText.includes('.mermaid-rendered')) {
                                markdownStyles += cssText + '\n';
                            }
                        }
                    } catch (e) {
                        // Gracefully skip stylesheets that can't be accessed
                        // (This should rarely happen now that we skip external stylesheets)
                        console.debug('Skipping stylesheet:', sheet.href);
                    }
                }
                
                // Create standalone HTML document with inline libraries (fully self-contained)
                // Fetch all required library files and inline them
                const [highlightJs, mathJaxJs, mermaidJs] = await Promise.all([
                    fetch('/static/libs/highlight.js/11.11.1/highlight.min.js').then(r => r.text()).catch(() => ''),
                    fetch('/static/libs/mathjax/3.2.2/es5/tex-mml-chtml.js').then(r => r.text()).catch(() => ''),
                    fetch('/static/libs/mermaid/11.12.2/dist/mermaid.min.js').then(r => r.text()).catch(() => '')
                ]);

                // Get highlight.js theme CSS
                let highlightThemeCss = '';
                if (highlightTheme) {
                    const themeName = highlightTheme.split('/').pop();
                    highlightThemeCss = await fetch(`/static/libs/highlight.js/11.11.1/styles/${themeName}`)
                        .then(r => r.text())
                        .catch(() => '');
                }

                const htmlDocument = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${noteName}</title>

    <!-- Highlight.js for code syntax highlighting (inline) -->
    <style>${highlightThemeCss}</style>
    <script>${highlightJs}<\/script>

    <!-- MathJax for LaTeX math rendering (inline) -->
    <script>
        MathJax = {
            tex: {
                inlineMath: [['$', '$']],
                displayMath: [['$$', '$$']],
                processEscapes: true,
                processEnvironments: true
            },
            options: {
                skipHtmlTags: ['script', 'noscript', 'style', 'textarea', 'pre', 'code']
            },
            startup: {
                pageReady: () => {
                    return MathJax.startup.defaultPageReady().then(() => {
                        // Highlight code blocks after MathJax is done (exclude diagram renderers)
                        document.querySelectorAll('pre code:not(.language-mermaid)').forEach((block) => {
                            hljs.highlightElement(block);
                        });
                    });
                }
            }
        };
    </script>
    <script>${mathJaxJs}<\/script>

    <!-- Mermaid.js for diagrams (inline) -->
    <script>${mermaidJs}<\/script>
    <script>
        const isDark = ${this.getThemeType() === 'dark'};
        mermaid.initialize({
            startOnLoad: false,
            theme: isDark ? 'dark' : 'default',
            securityLevel: 'strict',
            fontFamily: 'inherit',
            flowchart: { useMaxWidth: true },
            sequence: { useMaxWidth: true },
            gantt: { useMaxWidth: true },
            state: { useMaxWidth: true },
            er: { useMaxWidth: true },
            pie: { useMaxWidth: true },
            mindmap: { useMaxWidth: true },
            gitGraph: { useMaxWidth: true }
        });

        // Render any Mermaid code blocks
        document.addEventListener('DOMContentLoaded', async () => {
            const mermaidBlocks = document.querySelectorAll('pre code.language-mermaid');
            for (let i = 0; i < mermaidBlocks.length; i++) {
                const block = mermaidBlocks[i];
                const pre = block.parentElement;
                try {
                    const code = block.textContent;
                    const id = 'mermaid-diagram-' + i;
                    const { svg } = await window.mermaid.render(id, code);
                    const container = document.createElement('div');
                    container.className = 'mermaid-rendered';
                    container.style.cssText = 'background-color: transparent; padding: 20px; text-align: center; overflow-x: auto;';
                    container.innerHTML = svg;
                    pre.parentElement.replaceChild(container, pre);
                } catch (error) {
                    console.error('Mermaid rendering error:', error);
                }
            }
        });
    </script>
    
    <style>
        /* Theme CSS */
        ${themeCss}
        
        /* Base styles */
        * {
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            margin: 0;
            padding: 2rem;
            max-width: 900px;
            margin-left: auto;
            margin-right: auto;
            background-color: var(--bg-primary);
            color: var(--text-primary);
        }
        
        /* Markdown preview styles extracted from current page */
        ${markdownStyles}
        
        @media (max-width: 768px) {
            body {
                padding: 1rem;
            }
        }
        
        @media print {
            body {
                padding: 0.5in;
                max-width: none;
            }
        }
    </style>
</head>
<body>
    <div class="markdown-preview">
        ${renderedHTML}
    </div>
</body>
</html>`;
                
                // Create blob and download
                const blob = new Blob([htmlDocument], { type: 'text/html;charset=utf-8' });
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = `${noteName}.html`;
                document.body.appendChild(a);
                a.click();
                
                // Cleanup
                URL.revokeObjectURL(url);
                document.body.removeChild(a);
                
            } catch (error) {
                console.error('HTML export failed:', error);
                this.showAlert(this.t('export.failed', { error: error.message }));
            }
        },
        
        // Copy current note link to clipboard
        async copyNoteLink() {
            if (!this.note.current) return;
            
            // Build the full URL
            const pathWithoutExtension = this.note.current.replace('.md', '');
            const encodedPath = pathWithoutExtension.split('/').map(segment => encodeURIComponent(segment)).join('/');
            const url = `${window.location.origin}/${encodedPath}`;
            
            try {
                await navigator.clipboard.writeText(url);
            } catch (error) {
                // Fallback for older browsers
                const textArea = document.createElement('textarea');
                textArea.value = url;
                document.body.appendChild(textArea);
                textArea.select();
                document.execCommand('copy');
                document.body.removeChild(textArea);
            }
            
            // Show brief "Copied!" feedback
            this.ui.linkCopied = true;
            setTimeout(() => {
                this.ui.linkCopied = false;
            }, 1500);
        },
        
        // ============================================================================
        // Modal Helpers (replacements for native alert/confirm/prompt)
        // ============================================================================
        
        // Show a confirm dialog. Returns true if user confirmed, false otherwise.
        showConfirm(message) {
            return new Promise((resolve) => {
                this.modals.confirm.message = message;
                this.modals.confirm.resolve = resolve;
                this.modals.confirm.show = true;
            });
        },
        
        // Show an alert dialog (fire-and-forget, no return value)
        showAlert(message) {
            this.modals.alert.message = message;
            this.modals.alert.show = true;
        },
        
        // Show a prompt dialog. Returns the entered value, or null if cancelled.
        showPrompt(message, defaultValue = '') {
            return new Promise((resolve) => {
                this.modals.prompt.message = message;
                this.modals.prompt.value = defaultValue;
                this.modals.prompt.resolve = resolve;
                this.modals.prompt.show = true;
            });
        },
        
        // ============================================================================
        // Share Functions
        // ============================================================================
        
        // Load list of shared note paths (for visual indicators)
        async loadSharedNotePaths() {
            try {
                const response = await fetch('/api/shared-notes');
                if (response.ok) {
                    const data = await response.json();
                    NoteCache.sharedNotePaths = new Set(data.paths || []);
                }
            } catch (error) {
                console.error('Failed to load shared note paths:', error);
                NoteCache.sharedNotePaths = new Set();
            }
        },
        
        // Check if a note is currently shared (O(1) lookup)
        isNoteShared(notePath) {
            return NoteCache.sharedNotePaths.has(notePath);
        },
        
        // ============================================
        // Quick Switcher (Ctrl+Alt+P)
        // ============================================
        
        openQuickSwitcher() {
            this.modals.quickSwitcher.show = true;
            this.modals.quickSwitcher.query = '';
            this.modals.quickSwitcher.index = 0;
            // Populate initial results (only notes, not media files)
            const notesOnly = (this.notes || []).filter(n => n.type === 'note');
            this.modals.quickSwitcher.results = notesOnly.slice(0, 10);
            // Focus the input after the modal renders
            this.$nextTick(() => {
                const input = document.getElementById('quickSwitcherInput');
                if (input) input.focus();
            });
        },
        
        closeQuickSwitcher() {
            this.modals.quickSwitcher.show = false;
            this.modals.quickSwitcher.query = '';
            this.modals.quickSwitcher.index = 0;
        },
        
        // Filter notes for quick switcher based on query
        filterQuickSwitcher(query) {
            // Only include actual notes, not images
            const notes = (this.notes || []).filter(n => n.type === 'note');
            if (!query || !query.trim()) {
                // Show recent notes when no query
                return notes.slice(0, 10);
            }
            const q = query.toLowerCase();
            return notes
                .filter(n => 
                    n.name.toLowerCase().includes(q) || 
                    n.path.toLowerCase().includes(q)
                )
                .slice(0, 10);
        },
        
        // Handle keyboard navigation in quick switcher
        handleQuickSwitcherKeydown(e) {
            const results = this.modals.quickSwitcher.results;
            
            if (e.key === 'ArrowDown') {
                e.preventDefault();
                this.modals.quickSwitcher.index = Math.min(this.modals.quickSwitcher.index + 1, results.length - 1);
                this.scrollQuickSwitcherIntoView();
            } else if (e.key === 'ArrowUp') {
                e.preventDefault();
                this.modals.quickSwitcher.index = Math.max(this.modals.quickSwitcher.index - 1, 0);
                this.scrollQuickSwitcherIntoView();
            } else if (e.key === 'Enter') {
                e.preventDefault();
                const note = results[this.modals.quickSwitcher.index];
                if (note) {
                    this.loadNote(note.path);
                    this.closeQuickSwitcher();
                }
            } else if (e.key === 'Escape') {
                e.preventDefault();
                this.closeQuickSwitcher();
            }
        },
        
        // Scroll selected item into view in quick switcher
        scrollQuickSwitcherIntoView() {
            this.$nextTick(() => {
                const items = document.querySelectorAll('[data-quick-switcher-item]');
                if (items[this.modals.quickSwitcher.index]) {
                    items[this.modals.quickSwitcher.index].scrollIntoView({ block: 'nearest' });
                }
            });
        },
        
        // Select note from quick switcher by click
        selectQuickSwitcherNote(note) {
            this.loadNote(note.path);
            this.closeQuickSwitcher();
        },
        
        // Close share modal and reset state after animation
        closeShareModal() {
            this.modals.share.show = false;
            // Delay state reset until modal is fully hidden
            setTimeout(() => {
                this.modals.share.showQR = false;
                this.modals.share.info = null;
                this.modals.share.loading = false;
            }, 200);
        },
        
        // Generate QR code for share URL
        generateQRCode(url) {
            if (!url || typeof qrcode === 'undefined') return '';
            try {
                const qr = qrcode(0, 'M'); // 0 = auto version, M = medium error correction
                qr.addData(url);
                qr.make();
                return qr.createDataURL(4); // 4 = module size in pixels
            } catch (e) {
                console.error('QR code generation failed:', e);
                return '';
            }
        },
        
        // Open share modal and fetch current share status
        async openShareModal() {
            if (!this.note.current) return;
            
            // Reset state BEFORE showing modal to prevent flicker
            this.modals.share.showQR = false;
            this.modals.share.info = null;
            this.modals.share.loading = true;
            this.modals.share.show = true;
            
            try {
                const notePath = this.note.current.replace('.md', '');
                const encodedPath = notePath.split('/').map(segment => encodeURIComponent(segment)).join('/');
                const response = await fetch(`/api/share/${encodedPath}`);
                
                if (response.ok) {
                    this.modals.share.info = await response.json();
                } else {
                    this.modals.share.info = { shared: false };
                }
            } catch (error) {
                console.error('Failed to get share status:', error);
                this.modals.share.info = { shared: false };
            } finally {
                this.modals.share.loading = false;
            }
        },
        
        // Create a share link for the current note (with current theme)
        async createShareLink() {
            if (!this.note.current) return;
            
            this.modals.share.loading = true;
            
            try {
                const notePath = this.note.current.replace('.md', '');
                const encodedPath = notePath.split('/').map(segment => encodeURIComponent(segment)).join('/');
                const response = await secureFetch(`/api/share/${encodedPath}`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ theme: this.theme.current || 'light' })
                });
                
                if (response.ok) {
                    this.modals.share.info = await response.json();
                    this.modals.share.info.shared = true;
                    // Update the shared paths set
                    NoteCache.sharedNotePaths.add(this.note.current);
                } else {
                    const error = await response.json();
                    this.showAlert(this.t('share.error_creating', { error: error.detail || 'Unknown error' }));

                }
            } finally {
                this.modals.share.loading = false;
            }
        },
        
        // Copy share link to clipboard
        async copyShareLink() {
            if (!this.modals.share.info?.url) return;
            
            try {
                await navigator.clipboard.writeText(this.modals.share.info.url);
            } catch (error) {
                // Fallback for older browsers
                const textArea = document.createElement('textarea');
                textArea.value = this.modals.share.info.url;
                document.body.appendChild(textArea);
                textArea.select();
                document.execCommand('copy');
                document.body.removeChild(textArea);
            }
            
            this.modals.share.linkCopied = true;
            setTimeout(() => {
                this.modals.share.linkCopied = false;
            }, 2000);
        },
        
        // Revoke share link
        async revokeShareLink() {
            if (!this.note.current) return;
            
            if (!await this.showConfirm(this.t('share.confirm_revoke'))) return;
            
            this.modals.share.loading = true;
            
            try {
                const notePath = this.note.current.replace('.md', '');
                const encodedPath = notePath.split('/').map(segment => encodeURIComponent(segment)).join('/');
                const response = await secureFetch(`/api/share/${encodedPath}`, {
                    method: 'DELETE'
                });
                
                if (response.ok) {
                    this.modals.share.info = { shared: false };
                    // Update the shared paths set
                    NoteCache.sharedNotePaths.delete(this.note.current);
                } else {
                    const error = await response.json();
                    this.showAlert(this.t('share.error_revoking', { error: error.detail || 'Unknown error' }));

                }
                this.showAlert(this.t('share.error_revoking', { error: error.message }));
            } finally {
                this.modals.share.loading = false;
            }
        },
        
        // ============================================================================
        // Orphaned Media Cleanup Functions
        // ============================================================================
        
        // Show orphaned media cleanup modal and scan for orphaned files
        async openOrphanedMediaModal() {
            this.modals.orphanedMedia.show = true;
            this.modals.orphanedMedia.scanned = false;
            this.modals.orphanedMedia.files = [];
            this.modals.orphanedMedia.totalSize = 0;
            this.modals.orphanedMedia.error = null;
            this.modals.orphanedMedia.cleanupSuccess = false;
            await this.scanOrphanedMedia();
        },
        
        // Hide orphaned media cleanup modal
        closeOrphanedMediaModal() {
            this.modals.orphanedMedia.show = false;
            // Reset state after modal closes
            setTimeout(() => {
                this.modals.orphanedMedia.scanned = false;
                this.modals.orphanedMedia.files = [];
                this.modals.orphanedMedia.totalSize = 0;
                this.modals.orphanedMedia.error = null;
                this.modals.orphanedMedia.cleanupSuccess = false;
            }, 200);
        },
        
        // Scan for orphaned media files
        async scanOrphanedMedia() {
            this.modals.orphanedMedia.loading = true;
            this.modals.orphanedMedia.error = null;
            this.modals.orphanedMedia.cleanupSuccess = false;
            
            try {
                const response = await fetch('/api/media/orphaned');
                
                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(error.detail || 'Failed to scan for orphaned media');
                }
                
                const data = await response.json();
                this.modals.orphanedMedia.files = data.files || [];
                this.modals.orphanedMedia.totalSize = data.total_size || 0;
                this.modals.orphanedMedia.scanned = true;
            } catch (error) {
                ErrorHandler.handle('scan orphaned media', error);
                this.modals.orphanedMedia.error = error.message;
            } finally {
                this.modals.orphanedMedia.loading = false;
            }
        },
        
        // Delete orphaned media files
        async deleteAllOrphanedMedia() {
            if (this.modals.orphanedMedia.files.length === 0) return;
            
            const confirmed = await this.showConfirm(
                this.t('media.cleanup_confirm', { count: this.modals.orphanedMedia.files.length }) ||
                `Are you sure you want to delete ${this.modals.orphanedMedia.files.length} orphaned file(s)? This action cannot be undone.`
            );
            
            if (!confirmed) return;
            
            this.modals.orphanedMedia.cleanupInProgress = true;
            this.modals.orphanedMedia.error = null;
            
            try {
                const response = await secureFetch('/api/media/orphaned', {
                    method: 'DELETE'
                });
                
                if (!response.ok) {
                    const error = await response.json();
                    throw new Error(error.detail || 'Failed to cleanup orphaned media');
                }
                
                const data = await response.json();
                this.modals.orphanedMedia.cleanupSuccess = true;
                this.modals.orphanedMedia.files = [];
                this.modals.orphanedMedia.totalSize = 0;
                this.modals.orphanedMedia.scanned = false;
                
                // Refresh file list to show updated state
                await this.loadNotes();
                
                // Show success message
                const message = this.t('media.cleanup_success', { count: data.deletedCount }) ||
                    `Successfully deleted ${data.deletedCount} orphaned file(s).`;
                this.showAlert(message);
            } catch (error) {
                ErrorHandler.handle('cleanup orphaned media', error);
                this.modals.orphanedMedia.error = error.message;
            } finally {
                this.modals.orphanedMedia.cleanupInProgress = false;
            }
        },
        
        // Format file size in human-readable format
        formatFileSize(bytes) {
            if (bytes === 0) return '0 B';
            if (!bytes || isNaN(bytes)) return '0 B';
            
            const units = ['B', 'KB', 'MB', 'GB', 'TB'];
            const k = 1024;
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + units[i];
        },
        
        // Alias for formatFileSize (used in HTML templates)
        formatBytes(bytes) {
            return this.formatFileSize(bytes);
        },
        
        // Toggle Zen Mode (full immersive writing experience)
        async toggleZenMode() {
            if (!this.ui.zenMode) {
                // Entering Zen Mode
                this.ui.previousViewMode = this.ui.viewMode;
                this.ui.viewMode = 'edit';
                this.ui.mobileSidebarOpen = false;
                this.ui.zenMode = true;
                
                // Request fullscreen
                try {
                    const elem = document.documentElement;
                    if (elem.requestFullscreen) {
                        await elem.requestFullscreen();
                    } else if (elem.webkitRequestFullscreen) {
                        await elem.webkitRequestFullscreen();
                    } else if (elem.msRequestFullscreen) {
                        await elem.msRequestFullscreen();
                    }
                } catch (e) {
                    // Fullscreen not supported or denied, continue anyway
                    console.log('Fullscreen not available:', e);
                }
                
                // Focus editor after transition
                setTimeout(() => {
                    const editor = document.getElementById('note-editor');
                    if (editor) editor.focus();
                }, 300);
            } else {
                // Exiting Zen Mode
                this.ui.zenMode = false;
                this.ui.viewMode = this.ui.previousViewMode;
                
                // Exit fullscreen
                try {
                    if (document.exitFullscreen) {
                        await document.exitFullscreen();
                    } else if (document.webkitExitFullscreen) {
                        await document.webkitExitFullscreen();
                    } else if (document.msExitFullscreen) {
                        await document.msExitFullscreen();
                    }
                } catch (e) {
                    console.log('Exit fullscreen error:', e);
                }
            }
        },
        
        // Homepage folder navigation methods
        goToHomepageFolder(folderPath) {
            this.graph.show = false; // Close graph when navigating
            this.homepage.selectedFolder = folderPath || '';
            
            // Clear editor state to show landing page
            this.note.current = '';
            this.note.name = '';
            this.note.content = '';
            this.media.current = '';
            this.note.outline = [];
            document.title = this.app.name;
            
            // Invalidate cache to force recalculation
            NoteCache.resetHomepageCache();
            
            window.history.pushState({ homepageFolder: folderPath || '' }, '', '/');
        },
        
        // Navigate to homepage root and clear all editor state
        goHome() {
            this.graph.show = false; // Close graph when going home
            this.homepage.selectedFolder = '';
            this.note.current = '';
            this.note.name = '';
            this.note.content = '';
            this.media.current = '';
            this.note.outline = [];
            this.ui.mobileSidebarOpen = false;
            document.title = this.app.name;
            
            // Clear undo/redo history
            this.history.undo = [];
            this.history.redo = [];
            this.history.hasPendingChanges = false;
            
            // Invalidate cache to force recalculation
            NoteCache.resetHomepageCache();
            
            window.history.pushState({ homepageFolder: '' }, '', '/');
        },
        
        // Mobile files/home tab - context-aware behavior
        mobileFilesTabClick() {
            if (this.note.current || this.media.current || this.graph.show) {
                // Viewing content → go home
                this.goHome();
            } else {
                // On homepage → toggle files sidebar
                this.ui.activePanel = 'files';
                this.ui.mobileSidebarOpen = !this.ui.mobileSidebarOpen;
            }
        },
        
        // ==================== GRAPH VIEW ====================
        
        // Initialize the graph visualization
        async initGraph() {
            // Check if vis is loaded
            if (typeof vis === 'undefined') {
                console.error('vis-network library not loaded');
                return;
            }
            
            this.graph.loaded = false;
            
            try {
                // Fetch graph data from API
                const response = await fetch('/api/graph');
                if (!response.ok) throw new Error('Failed to fetch graph data');
                const data = await response.json();
                this.graph.data = data;
                
                // Get container
                const container = document.getElementById('graph-overlay');
                if (!container) return;
                
                // Get theme colors (force reflow to ensure CSS is applied)
                document.body.offsetHeight; // Force reflow
                const style = getComputedStyle(document.documentElement);
                
                // Helper to get CSS variable with fallback
                const getCssVar = (name, fallback) => {
                    const value = style.getPropertyValue(name).trim();
                    return value || fallback;
                };
                
                const accentPrimary = getCssVar('--accent-primary', '#7c3aed');
                const accentSecondary = getCssVar('--accent-secondary', '#a78bfa');
                const textPrimary = getCssVar('--text-primary', '#111827');
                const textSecondary = getCssVar('--text-secondary', '#6b7280');
                const bgPrimary = getCssVar('--bg-primary', '#ffffff');
                const bgSecondary = getCssVar('--bg-secondary', '#f3f4f6');
                const borderColor = getCssVar('--border-primary', '#e5e7eb');
                
                // Prepare nodes with styling - all nodes same base color
                const nodes = new vis.DataSet(data.nodes.map(n => ({
                    id: n.id,
                    label: n.label,
                    title: n.id, // Tooltip shows full path
                    color: {
                        background: accentPrimary,
                        border: accentPrimary,
                        highlight: {
                            background: accentPrimary,
                            border: textPrimary  // Darker border when selected
                        },
                        hover: {
                            background: accentSecondary,
                            border: accentPrimary
                        }
                    },
                    font: {
                        color: textPrimary,
                        size: 12,
                        face: 'system-ui, -apple-system, sans-serif'
                    },
                    borderWidth: this.note.current === n.id ? 4 : 2,
                    chosen: {
                        node: (values) => {
                            values.size = 22;
                            values.borderWidth = 4;
                            values.borderColor = textPrimary;
                        }
                    }
                })));
                
                // Prepare edges with styling based on type
                const edges = new vis.DataSet(data.edges.map((e, i) => ({
                    id: i,
                    from: e.source,
                    to: e.target,
                    color: {
                        color: e.type === 'wikilink' ? accentPrimary : borderColor,
                        highlight: accentPrimary,
                        hover: accentSecondary,
                        opacity: 0.8
                    },
                    width: e.type === 'wikilink' ? 2 : 1,
                    smooth: {
                        type: 'continuous',
                        roundness: 0.5
                    },
                    chosen: {
                        edge: (values) => {
                            values.width = 3;
                            values.color = accentPrimary;
                        }
                    }
                })));
                
                // Network options
                const options = {
                    nodes: {
                        shape: 'dot',
                        size: 16,
                        borderWidth: 2,
                        shadow: {
                            enabled: true,
                            color: 'rgba(0,0,0,0.1)',
                            size: 5,
                            x: 2,
                            y: 2
                        }
                    },
                    edges: {
                        arrows: {
                            to: {
                                enabled: true,
                                scaleFactor: 0.5,
                                type: 'arrow'
                            }
                        }
                    },
                    physics: {
                        enabled: true,
                        solver: 'forceAtlas2Based',
                        forceAtlas2Based: {
                            gravitationalConstant: -50,
                            centralGravity: 0.01,
                            springLength: 100,
                            springConstant: 0.08,
                            damping: 0.4,
                            avoidOverlap: 0.5
                        },
                        stabilization: {
                            enabled: true,
                            iterations: 200,
                            updateInterval: 25
                        }
                    },
                    interaction: {
                        hover: true,
                        tooltipDelay: 200,
                        navigationButtons: false,  // Using custom buttons instead
                        keyboard: {
                            enabled: true,
                            bindToWindow: false
                        },
                        zoomView: true,
                        dragView: true
                    },
                    layout: {
                        improvedLayout: true,
                        randomSeed: 42
                    }
                };
                
                // Destroy existing instance if any
                if (this.graph.instance) {
                    this.graph.instance.destroy();
                    this.graph.instance = null;
                }
                
                // Clear container to ensure clean state
                const graphCanvas = container.querySelector('canvas');
                if (graphCanvas) graphCanvas.remove();
                const visElements = container.querySelectorAll('.vis-network, .vis-navigation');
                visElements.forEach(el => el.remove());
                
                // Create the network
                this.graph.instance = new vis.Network(container, { nodes, edges }, options);
                
                // Store reference for callbacks
                const graphRef = this.graph.instance;
                const currentNoteRef = this.note.current;
                
                // Wait for stabilization
                this.graph.instance.once('stabilizationIterationsDone', () => {
                    graphRef.setOptions({ physics: { enabled: false } });
                    this.graph.loaded = true;
                    
                    // Focus and select current note if one is loaded
                    if (currentNoteRef) {
                        setTimeout(() => {
                            try {
                                if (graphRef && this.graph.show) {
                                    const nodeIds = graphRef.body.data.nodes.getIds();
                                    if (nodeIds.includes(currentNoteRef)) {
                                        // Focus on the node
                                        graphRef.focus(currentNoteRef, {
                                            scale: 1.2,
                                            animation: {
                                                duration: 500,
                                                easingFunction: 'easeInOutQuad'
                                            }
                                        });
                                        // Select the node to highlight it
                                        graphRef.selectNodes([currentNoteRef]);
                                    }
                                }
                            } catch (e) {
                                // Ignore - graph may have been destroyed
                            }
                        }, 150);
                    }
                });
                
                // Click event - open note
                this.graph.instance.on('click', (params) => {
                    if (params.nodes.length > 0) {
                        const noteId = params.nodes[0];
                        this.loadNote(noteId);
                        // Node is already selected by vis-network on click, no need to call selectNodes
                    }
                });
                
                // Double-click event - open note and close graph
                this.graph.instance.on('doubleClick', (params) => {
                    if (params.nodes.length > 0) {
                        const noteId = params.nodes[0];
                        // Close graph and load note
                        this.graph.show = false;
                        this.loadNote(noteId);
                    }
                });
                
                // Hover event - highlight connections
                this.graph.instance.on('hoverNode', (params) => {
                    const nodeId = params.node;
                    const connectedNodes = this.graph.instance.getConnectedNodes(nodeId);
                    const connectedEdges = this.graph.instance.getConnectedEdges(nodeId);
                    
                    // Dim all nodes except hovered and connected
                    const allNodes = nodes.getIds();
                    const updates = allNodes.map(id => ({
                        id,
                        opacity: (id === nodeId || connectedNodes.includes(id)) ? 1 : 0.2
                    }));
                    nodes.update(updates);
                });
                
                this.graph.instance.on('blurNode', () => {
                    // Reset all nodes to full opacity
                    const allNodes = nodes.getIds();
                    const updates = allNodes.map(id => ({ id, opacity: 1 }));
                    nodes.update(updates);
                });
                
                // Add legend to container
                this.addGraphLegend(container, accentPrimary, borderColor, textSecondary);
                
            } catch (error) {
                console.error('Failed to initialize graph:', error);
                this.graph.loaded = true; // Stop loading indicator
            }
        },
        
        // Add legend to graph container
        addGraphLegend(container, wikiColor, mdColor, textColor) {
            // Remove existing legend if any
            const existingLegend = container.querySelector('.graph-legend');
            if (existingLegend) existingLegend.remove();
            
            const legend = document.createElement('div');
            legend.className = 'graph-legend';
            legend.innerHTML = `
                <div class="graph-legend-item">
                    <span class="graph-legend-dot" style="background: ${wikiColor};"></span>
                    <span style="color: ${textColor};">${this.t('graph.wikilinks')}</span>
                </div>
                <div class="graph-legend-item">
                    <span class="graph-legend-dot" style="background: ${mdColor};"></span>
                    <span style="color: ${textColor};">${this.t('graph.markdown_links')}</span>
                </div>
                <div style="margin-top: 8px; font-size: 10px; color: ${textColor}; opacity: 0.7;">
                    ${this.t('graph.click_hint')}
                </div>
            `;
            container.appendChild(legend);
        },
        
        // Refresh graph when theme changes
        refreshGraph() {
            if (this.ui.viewMode === 'graph' && this.graph.instance) {
                this.initGraph();
            }
        },
        
        // Zoom graph in/out
        zoomGraph(scale) {
            if (!this.graph.instance) return;
            
            try {
                const currentScale = this.graph.instance.getScale();
                const newScale = currentScale * scale;
                
                // Clamp scale between 0.1 and 5
                const clampedScale = Math.max(0.1, Math.min(5, newScale));
                
                // Get current view position
                const viewPosition = this.graph.instance.getViewPosition();
                
                // Apply zoom centered on current view
                this.graph.instance.moveTo({
                    scale: clampedScale,
                    position: viewPosition,
                    animation: {
                        duration: 300,
                        easingFunction: 'easeInOutQuad'
                    }
                });
            } catch (e) {
                console.error('Zoom failed:', e);
            }
        },
        
        // Reset graph view to fit all nodes
        resetGraphView() {
            if (!this.graph.instance) return;
            
            try {
                // Fit all nodes with animation
                this.graph.instance.fit({
                    animation: {
                        duration: 500,
                        easingFunction: 'easeInOutQuad'
                    }
                });
            } catch (e) {
                console.error('Reset view failed:', e);
            }
        },
        
        // ==================== WEBSOCKET REAL-TIME UPDATES ====================
        
        // Start WebSocket connection for real-time updates
        // @param {boolean} resetAttempts - Whether to reset reconnect attempts (true for manual start, false for auto reconnect)
        startWebSocket(resetAttempts = true) {
            if (this.ws.connection && this.ws.connection.readyState === WebSocket.OPEN) return;
            
            // Reset reconnect attempts and disabled state when manually starting (e.g., page becomes visible again)
            if (resetAttempts) {
                this.ws.reconnectAttempts = 0;
                this.ws.disabled = false;
            }
            
            // Don't attempt connection if disabled (max reconnect attempts reached)
            if (this.ws.disabled) {
                return;
            }
            
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws`;
            
            try {
                this.ws.connection = new WebSocket(wsUrl);
                
                this.ws.connection.onopen = () => {
                    console.log('WebSocket connected');
                    this.ws.reconnectAttempts = 0;
                    this.ws.disabled = false;
                };
                
                this.ws.connection.onmessage = (event) => {
                    try {
                        const msg = JSON.parse(event.data);
                        if (msg.type === 'notes_updated') {
                            this.reloadNotesFromServer();
                        }
                    } catch (e) {
                        // Ignore parse errors
                    }
                };
                
                this.ws.connection.onclose = (event) => {
                    console.log('WebSocket disconnected');
                    this.ws.connection = null;
                    // Attempt reconnect with exponential backoff
                    this.scheduleWebSocketReconnect();
                };
                
                this.ws.connection.onerror = (error) => {
                    console.warn('WebSocket error:', error);
                };
            } catch (error) {
                console.warn('Failed to create WebSocket:', error);
                this.scheduleWebSocketReconnect();
            }
        },
        
        // Manually retry WebSocket connection (resets disabled state)
        retryWebSocket() {
            this.ws.disabled = false;
            this.ws.reconnectAttempts = 0;
            this.startWebSocket(true);
        },
        
        // Stop WebSocket connection
        stopWebSocket() {
            if (this.ws.reconnectTimeout) {
                clearTimeout(this.ws.reconnectTimeout);
                this.ws.reconnectTimeout = null;
            }
            if (this.ws.connection) {
                this.ws.connection.close();
                this.ws.connection = null;
            }
        },
        
        // Schedule WebSocket reconnection with exponential backoff
        scheduleWebSocketReconnect() {
            if (this.ws.reconnectTimeout) return;
            
            const MAX_RECONNECT_ATTEMPTS = 10;
            
            // Stop reconnecting after max attempts
            if (this.ws.reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
                console.warn('WebSocket: Max reconnect attempts reached, real-time sync temporarily disabled');
                this.ws.disabled = true;
                return;
            }
            
            // Exponential backoff: 1s, 2s, 4s, 8s, max 30s
            const delay = Math.min(1000 * Math.pow(2, this.ws.reconnectAttempts), 30000);
            this.ws.reconnectAttempts++;
            
            this.ws.reconnectTimeout = setTimeout(() => {
                this.ws.reconnectTimeout = null;
                this.startWebSocket(false); // Auto reconnect, don't reset attempts
            }, delay);
        },
        
        // Reload notes from server when notified via WebSocket
        async reloadNotesFromServer() {
            try {
                const response = await fetch('/api/notes?include_media=true');
                if (!response.ok) return;
                
                const data = await response.json();
                
                                const notesChanged = JSON.stringify(this.notes.map(n => n.path).sort()) !==
                                                    JSON.stringify(data.notes.map(n => n.path).sort());
                                const foldersChanged = JSON.stringify([...this.folders.all].sort()) !== 
                                                      JSON.stringify([...data.folders].sort());
                                
                                if (notesChanged || foldersChanged) {
                                    this.notes = data.notes;
                                    this.folders.all = data.folders;
                                    this.buildFolderTree();
                                }                
                this.ws.lastSyncTimestamp = Date.now();
            } catch (error) {
                // Silently fail
            }
        },
        
        // Legacy polling methods (kept as fallback)
        startPolling() {
            // Use WebSocket instead of polling
            this.startWebSocket();
        },
        
        stopPolling() {
            this.stopWebSocket();
        },
        
        async checkForUpdates() {
            // Delegate to WebSocket handler
            await this.reloadNotesFromServer();
        }
    }
}

