<template>
  <div id="app-container">
    <aside class="sidebar">
      <h1>GoBackup</h1>
      <nav>
        <a @click="currentView = 'backup'" :class="{active: currentView === 'backup'}">⚡ Backup</a>
        <a @click="currentView = 'restore'" :class="{active: currentView === 'restore'}">🔄 Restore</a>
      </nav>
    </aside>

    <main class="main-content">
      <!-- Backup View -->
      <div v-if="currentView === 'backup'" class="view">
        <h2>Create a New Backup</h2>
        <div class="card">
          <h3>Step 1: Select Paths</h3>
          <div class="input-group">
            <label>Source Directory:</label>
            <input v-model="backupSource" readonly type="text" placeholder="Select a directory to back up" />
            <button @click="selectSourceDir">Select</button>
          </div>
          <div class="input-group">
            <label>Destination Directory:</label>
            <input v-model="backupDest" readonly type="text" placeholder="Select a directory to save the backup" />
            <button @click="selectDestDir">Select</button>
          </div>
        </div>

        <div class="card">
          <h3>Step 2: Configure Filters (Optional)</h3>
          <div class="filter-grid">
            <!-- Name Filters -->
            <div class="filter-group">
              <label>Include Names (e.g. `*.log`, `data*`)</label>
              <textarea v-model="filters.includeNames" placeholder="One pattern per line"></textarea>
            </div>
            <div class="filter-group">
              <label>Exclude Names (e.g. `*.tmp`, `cache*`)</label>
              <textarea v-model="filters.excludeNames" placeholder="One pattern per line"></textarea>
            </div>
            <!-- Size Filters -->
            <div class="filter-group">
              <label>Min Size</label>
              <div class="size-input">
                <input type="number" v-model.number="filters.minSizeValue" min="0">
                <select v-model="filters.minSizeUnit">
                  <option>Bytes</option><option>KB</option><option>MB</option><option>GB</option>
                </select>
              </div>
            </div>
            <div class="filter-group">
              <label>Max Size</label>
              <div class="size-input">
                <input type="number" v-model.number="filters.maxSizeValue" min="0">
                <select v-model="filters.maxSizeUnit">
                  <option>Bytes</option><option>KB</option><option>MB</option><option>GB</option>
                </select>
              </div>
            </div>
            <!-- Date Filters -->
            <div class="filter-group">
              <label>Newer Than</label>
              <input type="date" v-model="filters.newerThan">
            </div>
            <div class="filter-group">
              <label>Older Than</label>
              <input type="date" v-model="filters.olderThan">
            </div>
          </div>
        </div>

        <div class="card">
          <h3>Step 3: Encryption (Optional)</h3>
          <div class="input-group">
            <label>Enable Encryption</label>
            <input type="checkbox" v-model="encryption.enabled" class="toggle"/>
          </div>
          <div v-if="encryption.enabled" class="encryption-options">
            <div class="input-group">
              <label>Password:</label>
              <input type="password" v-model="encryption.password" placeholder="Enter a strong password" />
            </div>
            <div class="input-group">
              <label>Algorithm:</label>
              <select v-model="encryption.algorithm">
                <option>AES-256</option>
                <option>ChaCha20</option>
              </select>
            </div>
          </div>
        </div>


        <div class="action-bar">
          <button class="primary" @click="doBackup" :disabled="!backupSource || !backupDest || inProgress">
            {{ inProgress ? 'Backing up...' : 'Start Backup' }}
          </button>
        </div>
      </div>

      <!-- Restore View -->
      <div v-if="currentView === 'restore'" class="view">
        <h2>Restore from a Backup</h2>
        <div class="card">
          <div class="input-group">
            <label>Backup File (.qbak):</label>
            <input v-model="restoreFile" readonly type="text" placeholder="Select a .qbak file to restore" />
            <button @click="selectRestoreFile">Select</button>
          </div>
          <div class="input-group">
            <label>Restore to Directory:</label>
            <input v-model="restoreDir" readonly type="text" placeholder="Select where to restore files" />
            <button @click="selectRestoreDir">Select</button>
          </div>
        </div>
        <div class="action-bar">
          <button class="primary" @click="doRestore()" :disabled="!restoreFile || !restoreDir || inProgress">
            {{ inProgress ? 'Restoring...' : 'Start Restore' }}
          </button>
        </div>
      </div>

      <!-- Log Section -->
      <div class="log-card">
        <h3>Log Output</h3>
        <p class="status">{{ statusMessage }}</p>
        <div class="log-box">
          <p v-for="(msg, index) in logMessages" :key="index">{{ msg }}</p>
        </div>
      </div>
    </main>

    <div v-if="isPasswordModalVisible" class="modal-overlay">
      <div class="modal-content">
        <h3>Enter Password</h3>
        <p>The backup file is encrypted. Please enter the password to continue.</p>
        <div class="input-group modal-input">
          <label>Password:</label>
          <input
              type="password"
              v-model="restorePasswordInput"
              @keyup.enter="submitPasswordAndRetryRestore"
              placeholder="Enter backup password"
              ref="passwordInputRef"
          />
        </div>
        <div class="modal-actions">
          <button @click="cancelPasswordPrompt">Cancel</button>
          <button class="primary" @click="submitPasswordAndRetryRestore">Submit</button>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import {ref, onMounted, reactive} from 'vue';
import { StartBackup, StartRestore, SelectDirectory, SelectFile } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

// --- State Management ---
const currentView = ref('backup');
const inProgress = ref(false);

// Backup State
const backupSource = ref('');
const backupDest = ref('');
const filters = ref({
  includeNames: '',
  excludeNames: '',
  minSizeValue: 0,
  minSizeUnit: 'Bytes',
  maxSizeValue: 0,
  maxSizeUnit: 'Bytes',
  newerThan: null,
  olderThan: null,
});
const encryption = reactive({ // Use reactive for the object
  enabled: false,
  password: '',
  algorithm: 'AES-256',
});

// Restore State
const restoreFile = ref('');
const restoreDir = ref('');

// Log State
const statusMessage = ref('Ready.');
const logMessages = ref([]);

const isPasswordModalVisible = ref(false);
const restorePasswordInput = ref('');
const passwordInputRef = ref(null); // 用于自动聚焦输入框

// --- Wails Event Listeners ---
onMounted(() => {
  const logHandler = (data) => {
    logMessages.value.unshift(data);
    if (logMessages.value.length > 200) logMessages.value.pop();
  };
  EventsOn("backup_progress", logHandler);
  EventsOn("restore_progress", logHandler);
});

// --- Helper Functions ---
const multipliers = { Bytes: 1, KB: 1024, MB: 1024**2, GB: 1024**3 };
const convertToBytes = (value, unit) => {
  if (!value || value <= 0) return 0;
  return value * (multipliers[unit] || 1);
};

const prepareFiltersForBackend = () => {
  const minSize = convertToBytes(filters.value.minSizeValue, filters.value.minSizeUnit);
  let maxSize = convertToBytes(filters.value.maxSizeValue, filters.value.maxSizeUnit);
  if (maxSize === 0) maxSize = -1; // Use -1 for no limit, as defined in Go backend

  return {
    includePaths: [], // TODO: Add UI for these if needed
    excludePaths: [], // TODO: Add UI for these if needed
    includeNames: filters.value.includeNames.split('\n').filter(s => s.trim() !== ''),
    excludeNames: filters.value.excludeNames.split('\n').filter(s => s.trim() !== ''),
    newerThan: filters.value.newerThan ? new Date(filters.value.newerThan).toISOString() : null,
    olderThan: filters.value.olderThan ? new Date(filters.value.olderThan).toISOString() : null,
    minSize: minSize,
    maxSize: maxSize,
  };
};

// --- UI Interaction Functions ---
const selectSourceDir = async () => { if(inProgress.value) return; const dir = await SelectDirectory(); if (dir) backupSource.value = dir; };
const selectDestDir = async () => { if(inProgress.value) return; const dir = await SelectDirectory(); if (dir) backupDest.value = dir; };
const selectRestoreFile = async () => { if(inProgress.value) return; const file = await SelectFile(); if (file) restoreFile.value = file; };
const selectRestoreDir = async () => { if(inProgress.value) return; const dir = await SelectDirectory(); if (dir) restoreDir.value = dir; };

// --- Core Actions ---
async function doBackup() {
  if (!backupSource.value || !backupDest.value) { /* ... */ return; }
  if (encryption.enabled && !encryption.password) {
    statusMessage.value = "Please enter a password for encryption.";
    return;
  }

  inProgress.value = true;
  statusMessage.value = "Preparing backup...";
  logMessages.value = [];

  const backendFilters = prepareFiltersForBackend();

  try {
    const result = await StartBackup({
      sourceDir: backupSource.value,
      destinationDir: backupDest.value,
      filters: backendFilters,
      // Pass encryption config
      useEncryption: encryption.enabled,
      encryptionAlgorithm: encryption.algorithm,
      encryptionPassword: encryption.password,
    });
    statusMessage.value = result;
  } catch (error) {
    statusMessage.value = `Error: ${error}`;
  } finally {
    inProgress.value = false;
  }
}


async function doRestore(password) {
  let cleanPassword = (typeof password === 'string') ? password : '';

  if (!restoreFile.value || !restoreDir.value) {
    statusMessage.value = "Please select backup file and restore directory.";
    return;
  }

  inProgress.value = true;
  statusMessage.value = "Restore in progress...";
  if (cleanPassword === '') {
    logMessages.value = [];
  }

  try {
    const result = await StartRestore({
      backupFile: restoreFile.value,
      restoreDir: restoreDir.value,
      password: cleanPassword,
    });
    statusMessage.value = result;
  } catch (error) {
    if (typeof error === 'string' && error.includes("password_required")) {
      statusMessage.value = "This backup is encrypted. Please provide the password.";
      // --- 修改点：不再使用 prompt，而是显示我们的模态框 ---
      isPasswordModalVisible.value = true;
      // 使用 nextTick 确保 DOM 更新后再聚焦
      await nextTick();
      passwordInputRef.value?.focus();

    } else {
      statusMessage.value = `Error: ${error}`;
    }
  } finally {
    // 只有在非密码请求的情况下才设置 inProgress=false
    // 如果弹出了密码框，则让它保持 inProgress 状态
    if (!isPasswordModalVisible.value) {
      inProgress.value = false;
    }
  }
}

// --- 新增：处理模态框的函数 ---

function submitPasswordAndRetryRestore() {
  if (!restorePasswordInput.value) {
    alert("Password cannot be empty.");
    return;
  }
  // 隐藏模态框
  isPasswordModalVisible.value = false;
  // 带着用户输入的密码重试恢复
  doRestore(restorePasswordInput.value);
  // 清空密码输入
  restorePasswordInput.value = '';
}

function cancelPasswordPrompt() {
  isPasswordModalVisible.value = false;
  statusMessage.value = "Password not provided. Restore cancelled.";
  inProgress.value = false; // 取消时，结束 inProgress 状态
  restorePasswordInput.value = '';
}
</script>

<style>
:root {
  --bg-color: #1a202c;
  --sidebar-bg: #2d3748;
  --card-bg: #2d3748;
  --text-color: #e2e8f0;
  --text-color-light: #a0aec0;
  --primary-color: #4299e1;
  --primary-color-hover: #2b6cb0;
  --border-color: #4a5568;
}

body, html {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 16px;
}

#app-container {
  display: flex;
  height: 100vh;
}

/* Sidebar */
.sidebar {
  width: 200px;
  background-color: var(--sidebar-bg);
  padding: 2rem 1rem;
  display: flex;
  flex-direction: column;
}
.sidebar h1 {
  font-size: 1.5rem;
  margin: 0 0 2rem 0;
  text-align: center;
}
.sidebar nav a {
  display: block;
  padding: 0.75rem 1rem;
  margin-bottom: 0.5rem;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
}
.sidebar nav a:hover {
  background-color: rgba(255, 255, 255, 0.05);
}
.sidebar nav a.active {
  background-color: var(--primary-color);
  font-weight: 500;
}

/* Main Content */
.main-content {
  flex-grow: 1;
  padding: 2rem;
  overflow-y: auto;
}
.view h2 {
  font-size: 1.8rem;
  font-weight: 600;
  margin-bottom: 1.5rem;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 1rem;
}
.card {
  background-color: var(--card-bg);
  border-radius: 8px;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}
.card h3 {
  margin-top: 0;
  font-size: 1.2rem;
  color: var(--text-color-light);
}

/* Forms & Inputs */
.input-group {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
}
.input-group label {
  width: 180px;
  text-align: right;
  margin-right: 1rem;
  color: var(--text-color-light);
}
.input-group input[type="text"],
.input-group input[type="date"] {
  flex-grow: 1;
  padding: 0.6rem;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  color: var(--text-color);
  border-radius: 4px;
}
textarea {
  width: 100%;
  min-height: 80px;
  padding: 0.6rem;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  color: var(--text-color);
  border-radius: 4px;
  resize: vertical;
}

/* Filters */
.filter-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}
.filter-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: var(--text-color-light);
  font-size: 0.9rem;
}
.size-input {
  display: flex;
}
.size-input input {
  flex-grow: 1;
  border-radius: 4px 0 0 4px;
  border-right: none;
}
.size-input select {
  padding: 0 0.5rem;
  background-color: var(--sidebar-bg);
  border: 1px solid var(--border-color);
  color: var(--text-color);
  border-radius: 0 4px 4px 0;
}

/* Buttons and Actions */
.action-bar {
  text-align: right;
}
button {
  padding: 0.6rem 1.2rem;
  background-color: var(--sidebar-bg);
  border: 1px solid var(--border-color);
  color: var(--text-color);
  border-radius: 6px;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.2s;
}
button:hover {
  border-color: var(--primary-color);
}
button.primary {
  background-color: var(--primary-color);
  border-color: var(--primary-color);
  font-weight: 500;
}
button.primary:hover {
  background-color: var(--primary-color-hover);
  border-color: var(--primary-color-hover);
}
button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Log Section */
.log-card {
  margin-top: 2rem;
}
.log-card h3 {
  margin: 0;
}
.status {
  color: var(--text-color-light);
  font-style: italic;
}
.log-box {
  height: 200px;
  overflow-y: auto;
  background-color: #000;
  border: 1px solid var(--border-color);
  text-align: left;
  padding: 10px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.9em;
  display: flex;
  flex-direction: column-reverse;
}

/* NEW/MODIFIED STYLES */
.encryption-options {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color);
}

/* A simple toggle switch for the checkbox */
input[type="checkbox"].toggle {
  -webkit-appearance: none;
  appearance: none;
  width: 40px;
  height: 22px;
  background-color: #4a5568;
  border-radius: 11px;
  position: relative;
  cursor: pointer;
  transition: background-color 0.2s;
}
input[type="checkbox"].toggle::before {
  content: '';
  position: absolute;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background-color: white;
  top: 2px;
  left: 2px;
  transition: transform 0.2s;
}
input[type="checkbox"].toggle:checked {
  background-color: var(--primary-color);
}
input[type="checkbox"].toggle:checked::before {
  transform: translateX(18px);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background-color: var(--sidebar-bg);
  padding: 2rem;
  border-radius: 8px;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 5px 15px rgba(0,0,0,0.3);
}

.modal-content h3 {
  margin-top: 0;
  font-size: 1.5rem;
}

.modal-content p {
  color: var(--text-color-light);
  margin-bottom: 1.5rem;
}

.modal-input {
  margin-bottom: 2rem;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
}
</style>