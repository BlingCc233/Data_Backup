<template>
  <div id="app-container">
    <!-- Main View: Home Screen -->

    <main class="main-content" v-if="currentScreen === 'home'">
      <div class="home-screen">
        <h1 class="home-title">GoBackup</h1>
        <div class="home-actions">
          <div class="action-card" @click="navigateTo('backup')">
            <div class="icon">⚡</div>
            <h2>备份</h2>
            <p>创建新的数据备份</p>
          </div>
          <div class="action-card" @click="navigateTo('restore')">
            <div class="icon">🔄</div>
            <h2>恢复</h2>
            <p>从备份文件中恢复数据</p>
          </div>
        </div>
      </div>
    </main>

    <!-- Nested Views: Backup/Restore pages -->
    <main class="main-content" v-else>
      <div class="view-header">
        <button class="back-button" @click="navigateBack">← 返回</button>
        <h2 class="view-title">{{ viewTitle }}</h2>
      </div>

      <!-- Backup View -->
      <div v-if="currentScreen === 'backup'" class="view">
        <!-- Step 1: Select Paths & Files -->
        <div v-if="backupStep === 1">
          <div class="card">
            <h3>Step 1: 选择要备份的文件或文件夹</h3>
            <p class="description">您可以多选文件或文件夹。选择后，它们将显示在下面的列表中，您可以进一步筛选。</p>
            <div class="action-bar-left">
              <button @click="selectBackupSources('files')">选择文件</button>
              <button @click="selectBackupSources('dirs')">选择文件夹</button>
            </div>
          </div>

          <div class="card file-list-card" v-if="backupFiles.length > 0">
            <div class="file-list-header">
              <h3>已选择的项目</h3>
              <span>共 {{ backupFiles.length }} 项</span>
            </div>
            <table class="file-table">
              <thead>
              <tr>
                <th class="col-checkbox"><input type="checkbox" @change="toggleSelectAllBackupFiles" :checked="allBackupFilesSelected"></th>
                <th @click="sortBackupFiles('name')" class="sortable">名称 <span v-if="sort.key === 'name'">{{ sort.order === 'asc' ? '▲' : '▼' }}</span></th>
                <th @click="sortBackupFiles('size')" class="sortable">大小 <span v-if="sort.key === 'size'">{{ sort.order === 'asc' ? '▲' : '▼' }}</span></th>
                <th @click="sortBackupFiles('modTime')" class="sortable">修改时间 <span v-if="sort.key === 'modTime'">{{ sort.order === 'asc' ? '▲' : '▼' }}</span></th>
                <th>权限</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="file in sortedBackupFiles" :key="file.path">
                <td><input type="checkbox" v-model="file.selected"></td>
                <td class="col-name">
                  <span class="file-icon">{{ file.isDir ? '📁' : '📄' }}</span>
                  {{ file.name }}
                  <small v-if="file.isDir" class="dir-hint">(点击文件夹可浏览，但此处仅作展示)</small>
                </td>
                <td>{{ formatSize(file.size) }}</td>
                <td>{{ formatDate(file.modTime) }}</td>
                <td>{{ file.mode }}</td>
              </tr>
              </tbody>
            </table>
          </div>

          <div class="action-bar">
            <button class="primary" @click="backupStep = 2" :disabled="selectedBackupFileCount === 0">
              下一步 ({{ selectedBackupFileCount }} 项)
            </button>
          </div>
        </div>

        <!-- Step 2: Advanced Filters -->
        <div v-if="backupStep === 2">
          <div class="card">
            <h3>Step 2: 高级筛选 (可选)</h3>
            <p class="description">
              您已手动选择了 {{ selectedBackupFileCount }} 个项目。在这里，您可以应用更高级的规则来进一步过滤这些项目。
              例如，在选中的文件夹中排除所有 `.tmp` 文件。
            </p>
            <div class="filter-grid">
              <!-- Name Filters -->
              <div class="filter-group">
                <label>包含名称 (e.g. `*.log`, `data*`)</label>
                <textarea v-model="filters.includeNames" placeholder="一行一个匹配模式"></textarea>
              </div>
              <div class="filter-group">
                <label>排除名称 (e.g. `*.tmp`, `cache*`)</label>
                <textarea v-model="filters.excludeNames" placeholder="一行一个匹配模式"></textarea>
              </div>
              <!-- Size Filters -->
              <div class="filter-group">
                <label>最小大小</label>
                <div class="size-input">
                  <input type="number" v-model.number="filters.minSizeValue" min="0">
                  <select v-model="filters.minSizeUnit">
                    <option>Bytes</option><option>KB</option><option>MB</option><option>GB</option>
                  </select>
                </div>
              </div>
              <div class="filter-group">
                <label>最大大小</label>
                <div class="size-input">
                  <input type="number" v-model.number="filters.maxSizeValue" min="0">
                  <select v-model="filters.maxSizeUnit">
                    <option>Bytes</option><option>KB</option><option>MB</option><option>GB</option>
                  </select>
                </div>
              </div>
              <!-- Date Filters -->
              <div class="filter-group">
                <label>晚于此日期</label>
                <input type="date" v-model="filters.newerThan">
              </div>
              <div class="filter-group">
                <label>早于此日期</label>
                <input type="date" v-model="filters.olderThan">
              </div>
            </div>
          </div>
          <div class="action-bar">
            <button @click="backupStep = 1">上一步</button>
            <button class="primary" @click="backupStep = 3">下一步</button>
          </div>
        </div>

        <!-- Step 3: Encryption & Start -->
        <div v-if="backupStep === 3">
          <div v-if="!inProgress">
            <div class="card">
              <h3>Step 3: 加密与执行</h3>
              <div class="input-group">
                <label>启用加密</label>
                <input type="checkbox" v-model="encryption.enabled" class="toggle"/>
              </div>
              <div v-if="encryption.enabled" class="encryption-options">
                <p class="description">
                  <strong>AES-256:</strong> 工业标准，安全可靠，硬件加速下性能优异。<br>
                  <strong>ChaCha20:</strong> 现代流式加密，在没有硬件加速的CPU上通常比AES更快。<br>
                  <strong style="color: #f6ad55;">注意: 启用加密会显著增加备份和恢复所需的时间。</strong>
                </p>
                <div class="input-group">
                  <label>密码:</label>
                  <input type="password" v-model="encryption.password" placeholder="输入一个强密码" />
                </div>
                <div class="input-group">
                  <label>算法:</label>
                  <select v-model="encryption.algorithm">
                    <option>AES-256</option>
                    <option>ChaCha20</option>
                  </select>
                </div>
              </div>
            </div>

            <div class="card">
              <h3>目标位置</h3>
              <div class="input-group">
                <label>备份到:</label>
                <input v-model="backupDest" readonly type="text" placeholder="选择一个目录来保存备份文件" />
                <button @click="selectDestDir">选择</button>
              </div>
            </div>


            <div class="action-bar">
              <button @click="backupStep = 2">上一步</button>
              <button class="primary" @click="doBackup" :disabled="!backupDest || (encryption.enabled && !encryption.password)">
                开始备份
              </button>
            </div>
          </div>
          <div v-else class="progress-view">
            <h3>正在备份...</h3>
            <progress :value="progress.value" :max="progress.max"></progress>
            <p class="status">{{ statusMessage }}</p>
            <div class="action-bar">
              <button class="danger" @click="stopOperation">停止备份</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Restore View -->
      <div v-if="currentScreen === 'restore'" class="view">
        <!-- Step 1: Select Backup -->
        <div v-if="restoreStep === 1">
          <div class="card">
            <h3>Step 1: 选择一个备份记录</h3>
            <p class="description">这里显示了您最近的备份记录。点击一项来查看其内容并进行恢复。</p>
            <div class="backup-history-list">
              <div v-for="item in backupHistory" :key="item.ID" class="history-item" @click="selectBackupFromHistory(item)">
                <div class="history-item-main">
                  <span class="history-file">{{ item.FileName }}</span>
                  <span class="history-date">{{ formatDate(item.CreatedAt) }}</span>
                </div>
                <small class="history-path">{{ item.BackupPath }}</small>
              </div>
              <div v-if="!backupHistory.length" class="history-empty">
                暂无备份记录。
              </div>
            </div>
          </div>
          <div class="card">
            <h3>或者, 手动选择备份文件</h3>
            <div class="input-group">
              <label>备份文件 (.qbak):</label>
              <input v-model="restoreFile" readonly type="text" placeholder="选择一个 .qbak 文件" />
              <button @click="selectRestoreFileManually">选择</button>
            </div>
          </div>
        </div>

        <!-- Step 2: Select files to restore / Enter Password -->
        <div v-if="restoreStep === 2">
          <div class="card">
            <h3>Step 2: 选择要恢复的内容</h3>
            <p class="description">您选择的备份文件 <strong>{{ selectedRestoreFile.name }}</strong> 是加密的。请输入密码以查看内容并继续。</p>
            <div class="input-group">
              <label>密码:</label>
              <input type="password" v-model="restorePasswordInput" @keyup.enter="listRestoreFiles" placeholder="输入备份密码" />
            </div>
          </div>
          <div class="action-bar">
            <button @click="restoreStep = 1">上一步</button>
            <button class="primary" @click="listRestoreFiles" :disabled="!restorePasswordInput">查看文件</button>
          </div>
        </div>

        <!-- Step 3: File List & Destination -->
        <div v-if="restoreStep === 3">
          <div class="card file-list-card">
            <div class="file-list-header">
              <h3>备份内容 ({{ selectedRestoreFile.name }})</h3>
              <span>共 {{ restoreFiles.length }} 项</span>
            </div>
            <table class="file-table">
              <thead>
              <tr>
                <th class="col-checkbox"><input type="checkbox" @change="toggleSelectAllRestoreFiles" :checked="allRestoreFilesSelected"></th>
                <th>名称</th>
                <th>大小</th>
                <th>修改时间</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="file in restoreFiles" :key="file.path">
                <td><input type="checkbox" v-model="file.selected"></td>
                <td class="col-name">
                  <span class="file-icon">{{ file.isDir ? '📁' : '📄' }}</span>
                  {{ file.path }}
                </td>
                <td>{{ formatSize(file.size) }}</td>
                <td>{{ formatDate(file.modTime) }}</td>
              </tr>
              </tbody>
            </table>
          </div>
          <div class="card">
            <h3>恢复到...</h3>
            <div class="input-group">
              <label>目标目录:</label>
              <input v-model="restoreDir" readonly type="text" placeholder="选择恢复文件的位置" />
              <button @click="selectRestoreDir">选择</button>
            </div>
          </div>
          <div class="action-bar">
            <button @click="navigateBack">上一步</button>
            <button class="primary" @click="doRestore()" :disabled="!restoreDir || selectedRestoreFileCount === 0">
              开始恢复 ({{ selectedRestoreFileCount }} 项)
            </button>
          </div>
        </div>

        <!-- Step 4: Restore Progress -->
        <div v-if="restoreStep === 4">
          <div class="progress-view">
            <h3>正在恢复...</h3>
            <progress :value="progress.value" :max="progress.max"></progress>
            <p class="status">{{ statusMessage }}</p>
            <div class="action-bar">
              <button class="danger" @click="stopOperation">停止恢复</button>
            </div>
          </div>
        </div>
      </div>

    </main>

    <!-- Global Log Section -->
    <div class="log-card-container" v-if="currentScreen !== 'home'">
      <div class="log-card">
        <h3>日志输出</h3>
        <p class="status">{{ statusMessage }}</p>
        <div class="log-box">
          <p v-for="(msg, index) in logMessages" :key="index">{{ msg }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import {ref, onMounted, reactive, computed, watch} from 'vue';
import {
  SelectFiles, SelectDirectory, GetFileMetadata, StartBackup, StopOperation,
  GetBackupHistory, ListBackupContents, StartRestore, OpenInExplorer
} from '../wailsjs/go/main/App';
import { EventsOn, EventsOnce } from '../wailsjs/runtime/runtime';

// --- Global State ---
const currentScreen = ref('home'); // 'home', 'backup', 'restore'
const historyStack = ref(['home']);
const inProgress = ref(false);
const statusMessage = ref('准备就绪。');
const logMessages = ref([]);
const progress = reactive({ value: 0, max: 100 });

// --- Navigation ---
const viewTitle = computed(() => {
  if (currentScreen.value === 'backup') return '创建新备份';
  if (currentScreen.value === 'restore') return '从备份恢复';
  return 'GoBackup';
});

function navigateTo(screen) {
  currentScreen.value = screen;
  historyStack.value.push(screen);
  // Reset states when entering a new screen
  if(screen === 'backup') resetBackupState();
  if(screen === 'restore') resetRestoreState();
}

function navigateBack() {
  currentScreen.value = 'home';
  historyStack.value = ['home']; // 重置历史栈，只保留首页
  // 重置备份和恢复状态
  resetBackupState();
  resetRestoreState();
}


// --- Backup State & Logic ---
const backupStep = ref(1);
const backupFiles = ref([]);
const backupDest = ref('');
const sort = reactive({ key: 'name', order: 'asc' });

const filters = ref({
  includeNames: '', excludeNames: '', minSizeValue: 0, minSizeUnit: 'Bytes',
  maxSizeValue: 0, maxSizeUnit: 'Bytes', newerThan: null, olderThan: null,
});
const encryption = reactive({ enabled: false, password: '', algorithm: 'AES-256' });

function resetBackupState() {
  backupStep.value = 1;
  backupFiles.value = [];
  backupDest.value = '';
  // reset filters and encryption if needed
}

async function selectBackupSources(type) {
  try {
    const paths = await SelectFiles(type === 'dirs');
    if (paths && paths.length > 0) {
      const metadata = await GetFileMetadata(paths);
      // Add 'selected' property and prevent duplicates
      const existingPaths = new Set(backupFiles.value.map(f => f.path));
      metadata.forEach(m => {
        if (!existingPaths.has(m.path)) {
          backupFiles.value.push({ ...m, selected: true });
        }
      });
    }
  } catch (error) {
    statusMessage.value = `Error selecting sources: ${error}`;
  }
}
const selectDestDir = async () => { if(inProgress.value) return; const dir = await SelectDirectory(); if (dir) backupDest.value = dir; };

const sortedBackupFiles = computed(() => {
  return [...backupFiles.value].sort((a, b) => {
    let modifier = sort.order === 'asc' ? 1 : -1;
    if (a[sort.key] < b[sort.key]) return -1 * modifier;
    if (a[sort.key] > b[sort.key]) return 1 * modifier;
    return 0;
  });
});

function sortBackupFiles(key) {
  if (sort.key === key) {
    sort.order = sort.order === 'asc' ? 'desc' : 'asc';
  } else {
    sort.key = key;
    sort.order = 'asc';
  }
}

const allBackupFilesSelected = computed(() => backupFiles.value.length > 0 && backupFiles.value.every(f => f.selected));
const selectedBackupFileCount = computed(() => backupFiles.value.filter(f => f.selected).length);

function toggleSelectAllBackupFiles(event) {
  const isChecked = event.target.checked;
  backupFiles.value.forEach(f => f.selected = isChecked);
}

async function doBackup() {
  const selectedPaths = backupFiles.value.filter(f => f.selected).map(f => f.path);
  if (selectedPaths.length === 0 || !backupDest.value) {
    statusMessage.value = "请选择要备份的文件和目标目录。";
    return;
  }
  if (encryption.enabled && !encryption.password) {
    statusMessage.value = "启用加密后，请输入密码。";
    return;
  }

  inProgress.value = true;
  statusMessage.value = "准备备份...";
  logMessages.value = [];
  progress.value = 0;
  progress.max = selectedPaths.length; // Approximate max value

  const multipliers = { Bytes: 1, KB: 1024, MB: 1024**2, GB: 1024**3 };
  const convertToBytes = (value, unit) => (value || 0) * (multipliers[unit] || 1);
  const minSize = convertToBytes(filters.value.minSizeValue, filters.value.minSizeUnit);
  let maxSize = convertToBytes(filters.value.maxSizeValue, filters.value.maxSizeUnit);
  if (maxSize <= 0) maxSize = -1;

  try {
    const result = await StartBackup({
      sourcePaths: selectedPaths,
      destinationDir: backupDest.value,
      filters: {
        includeNames: filters.value.includeNames.split('\n').filter(s => s.trim()),
        excludeNames: filters.value.excludeNames.split('\n').filter(s => s.trim()),
        newerThan: filters.value.newerThan ? new Date(filters.value.newerThan).toISOString() : null,
        olderThan: filters.value.olderThan ? new Date(filters.value.olderThan).toISOString() : null,
        minSize: minSize,
        maxSize: maxSize,
      },
      UseCompression: true,
      useEncryption: encryption.enabled,
      encryptionAlgorithm: encryption.algorithm,
      encryptionPassword: encryption.password,
    });
    statusMessage.value = result;
  } catch (error) {
    statusMessage.value = `错误: ${error}`;
  } finally {
    inProgress.value = false;
    await fetchBackupHistory();
  }
}

// --- Restore State & Logic ---
const restoreStep = ref(1);
const backupHistory = ref([]);
const restoreFile = ref(''); // manual selection path
const selectedRestoreFile = ref({ name: '', path: '', isEncrypted: false });
const restoreDir = ref('');
const restorePasswordInput = ref('');
const restoreFiles = ref([]); // files inside backup

function resetRestoreState() {
  restoreStep.value = 1;
  restoreFile.value = '';
  selectedRestoreFile.value = { name: '', path: '', isEncrypted: false };
  restoreDir.value = '';
  restorePasswordInput.value = '';
  restoreFiles.value = [];
  fetchBackupHistory();
}

async function fetchBackupHistory() {
  try {
    backupHistory.value = await GetBackupHistory();
  } catch(e) {
    statusMessage.value = `无法加载备份历史: ${e}`;
  }
}

function selectBackupFromHistory(item) {
  selectedRestoreFile.value.name = item.FileName;
  selectedRestoreFile.value.path = item.BackupPath;
  listRestoreFiles();
}

async function selectRestoreFileManually() {
  const path = await SelectFiles(false); // false for file
  if (path && path.length > 0) {
    restoreFile.value = path[0];
    selectedRestoreFile.value.path = path[0];
    selectedRestoreFile.value.name = path[0].split(/[/\\]/).pop();
    listRestoreFiles();
  }
}

async function listRestoreFiles() {
  if (!selectedRestoreFile.value.path) return;
  statusMessage.value = "正在读取备份文件内容...";
  try {
    const result = await ListBackupContents(selectedRestoreFile.value.path, restorePasswordInput.value);
    selectedRestoreFile.value.isEncrypted = false; // success, not encrypted or password correct
    restoreFiles.value = result.map(f => ({...f, selected: true}));
    restoreStep.value = 3;
  } catch (error) {
    if (error.includes("password_required")) {
      selectedRestoreFile.value.isEncrypted = true;
      restoreStep.value = 2; // Go to password page
      statusMessage.value = "此备份已加密，请输入密码。";
    } else {
      statusMessage.value = `无法读取备份文件: ${error}`;
      // Go back to selection
      restoreStep.value = 1;
    }
  }
}

const allRestoreFilesSelected = computed(() => restoreFiles.value.length > 0 && restoreFiles.value.every(f => f.selected));
const selectedRestoreFileCount = computed(() => restoreFiles.value.filter(f => f.selected).length);

function toggleSelectAllRestoreFiles(event) {
  const isChecked = event.target.checked;
  restoreFiles.value.forEach(f => f.selected = isChecked);
}
const selectRestoreDir = async () => { if(inProgress.value) return; const dir = await SelectDirectory(); if (dir) restoreDir.value = dir; };

async function doRestore() {
  const filesToRestore = restoreFiles.value.filter(f => f.selected).map(f => f.path);
  if (!selectedRestoreFile.value.path || !restoreDir.value || filesToRestore.length === 0) {
    statusMessage.value = "请选择备份文件、要恢复的文件和目标目录。";
    return;
  }

  inProgress.value = true;
  restoreStep.value = 4;
  statusMessage.value = "准备恢复...";
  logMessages.value = [];
  progress.value = 0;
  progress.max = filesToRestore.length;

  try {
    const result = await StartRestore({
      backupFile: selectedRestoreFile.value.path,
      restoreDir: restoreDir.value,
      password: restorePasswordInput.value,
      filesToRestore: filesToRestore,
    });
    statusMessage.value = result;
    if (result.includes("成功")) {
      await OpenInExplorer(restoreDir.value);
    }
  } catch (error) {
    statusMessage.value = `错误: ${error}`;
  } finally {
    inProgress.value = false;
    // Optionally go back to the list
    // restoreStep.value = 3;
  }
}

// --- Common & Helpers ---
function stopOperation() {
  StopOperation();
  inProgress.value = false;
  statusMessage.value = "操作已停止。";
  // Rollback to previous step
  if (currentScreen.value === 'backup' && backupStep.value === 3) backupStep.value = 2;
  if (currentScreen.value === 'restore' && restoreStep.value === 4) restoreStep.value = 3;
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const formatDate = (dateString) => {
  if (!dateString) return 'N/A';
  return new Date(dateString).toLocaleString();
};

onMounted(() => {
  EventsOn("log_message", (data) => {
    logMessages.value.unshift(data);
    if (logMessages.value.length > 200) logMessages.value.pop();
  });
  EventsOn("progress_update", (p) => {
    statusMessage.value = p.message;
    progress.value = p.current;
    if (p.total > 0) progress.max = p.total;
  });

  fetchBackupHistory();
});
</script>

<style>
/* Keeping original styles and adding new ones */
:root {
  --bg-color: #1a202c;
  --sidebar-bg: #2d3748; /* Re-purposed for cards */
  --card-bg: #2d3748;
  --text-color: #e2e8f0;
  --text-color-light: #a0aec0;
  --primary-color: #4299e1;
  --primary-color-hover: #2b6cb0;
  --border-color: #4a5568;
  --danger-color: #e53e3e;
}

body, html {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 16px;
  overflow: hidden;
}

#app-container {
  display: flex;
  flex-direction: column; /* Changed for new layout */
  height: 100vh;
}

/* Main Content Area */
.main-content {
  flex-grow: 1;
  padding: 2rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.view h2, .view-title {
  font-size: 1.8rem;
  font-weight: 600;
  margin-bottom: 0;
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
  color: var(--text-color);
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 0.75rem;
  margin-bottom: 1rem;
}
.description {
  color: var(--text-color-light);
  font-size: 0.95rem;
  line-height: 1.5;
  margin-top: -0.5rem;
  margin-bottom: 1.5rem;
}

/* Home Screen */
.home-screen {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  text-align: center;
}
.home-title {
  font-size: 3rem;
  font-weight: 700;
  margin-bottom: 2rem;
}
.home-actions {
  display: flex;
  gap: 2rem;
}
.action-card {
  background-color: var(--card-bg);
  padding: 2rem 3rem;
  border-radius: 12px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  border: 1px solid var(--border-color);
}
.action-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0,0,0,0.2);
  border-color: var(--primary-color);
}
.action-card .icon {
  font-size: 4rem;
  line-height: 1;
}
.action-card h2 {
  font-size: 1.5rem;
  margin: 1rem 0 0.5rem;
}
.action-card p {
  color: var(--text-color-light);
  margin: 0;
}

/* Nested View Header */
.view-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding-bottom: 1rem;
  margin-bottom: 1.5rem;
  border-bottom: 1px solid var(--border-color);
}
.back-button {
  background: none;
  border: 1px solid var(--border-color);
  color: var(--text-color-light);
  font-size: 1rem;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  cursor: pointer;
}
.back-button:hover {
  background-color: var(--card-bg);
  color: var(--text-color);
}


/* Forms & Inputs */
.input-group {
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
}
.input-group label {
  width: 120px;
  text-align: right;
  margin-right: 1rem;
  color: var(--text-color-light);
  flex-shrink: 0;
}
.input-group input[type="text"],
.input-group input[type="password"],
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

/* File List Table */
.file-list-card { padding: 0; }
.file-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--border-color);
}
.file-list-header h3 {
  border: none;
  padding: 0;
  margin: 0;
}
.file-table {
  width: 100%;
  border-collapse: collapse;
}
.file-table th, .file-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--border-color);
}
.file-table tbody tr:last-child td { border-bottom: none; }
.file-table th {
  color: var(--text-color-light);
  font-size: 0.9rem;
  text-transform: uppercase;
}
.file-table th.sortable { cursor: pointer; }
.file-table th.sortable:hover { color: var(--text-color); }
.file-table .col-checkbox { width: 40px; text-align: center; }
.file-table .col-name {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}
.file-icon { font-size: 1.2rem; }
.dir-hint { color: var(--text-color-light); margin-left: auto; }

/* Backup History */
.backup-history-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.history-item {
  background-color: var(--bg-color);
  padding: 0.75rem 1.25rem;
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid var(--border-color);
  transition: background-color 0.2s, border-color 0.2s;
}
.history-item:hover {
  background-color: rgba(66, 153, 225, 0.1);
  border-color: var(--primary-color);
}
.history-item-main {
  display: flex;
  justify-content: space-between;
  font-weight: 500;
}
.history-date { color: var(--text-color-light); font-size: 0.9rem; }
.history-path { color: var(--text-color-light); font-size: 0.8rem; }
.history-empty { text-align: center; padding: 2rem; color: var(--text-color-light); }

/* Progress View */
.progress-view {
  text-align: center;
  padding: 2rem;
}
.progress-view h3 { margin-top: 0; }
progress {
  width: 100%;
  height: 12px;
  -webkit-appearance: none;
  appearance: none;
  border: none;
  border-radius: 6px;
  overflow: hidden;
  background-color: var(--bg-color);
}
progress::-webkit-progress-bar { background-color: var(--bg-color); }
progress::-webkit-progress-value { background-color: var(--primary-color); transition: width 0.1s linear; }


/* Actions */
.action-bar { text-align: right; margin-top: 1.5rem; display: flex; justify-content: flex-end; gap: 1rem; }
.action-bar-left { text-align: left; display: flex; gap: 1rem; }
button {
  padding: 0.6rem 1.2rem; background-color: var(--sidebar-bg);
  border: 1px solid var(--border-color); color: var(--text-color);
  border-radius: 6px; cursor: pointer; font-size: 1rem; transition: all 0.2s;
}
button:hover { border-color: var(--primary-color); }
button.primary { background-color: var(--primary-color); border-color: var(--primary-color); font-weight: 500; }
button.primary:hover { background-color: var(--primary-color-hover); border-color: var(--primary-color-hover); }
button.danger { background-color: var(--danger-color); border-color: var(--danger-color); }
button:disabled { opacity: 0.5; cursor: not-allowed; }

/* Log Section (at the bottom) */
.log-card-container {
  flex-shrink: 0;
  padding: 0 2rem 2rem 2rem;
}
.log-card {
  margin-top: 0;
  background-color: var(--card-bg);
  border-radius: 8px;
  padding: 1rem 1.5rem;
}
.log-card h3 { margin: 0; font-size: 1rem; color: var(--text-color-light); }
.status { color: var(--text-color-light); font-style: italic; font-size: 0.9rem; }
.log-box {
  height: 150px; overflow-y: auto; background-color: #000;
  border: 1px solid var(--border-color); text-align: left; padding: 10px;
  border-radius: 4px; font-family: monospace; font-size: 0.9em;
  display: flex; flex-direction: column-reverse;
}

/* Other component styles (from original) */
.filter-grid { grid-template-columns: 1fr 1fr; gap: 1.5rem; display: grid; }
.filter-group label { display: block; margin-bottom: 0.5rem; color: var(--text-color-light); font-size: 0.9rem; }
.size-input { display: flex; }
.size-input input { flex-grow: 1; border-radius: 4px 0 0 4px; border-right: none; }
.size-input select { padding: 0 0.5rem; background-color: var(--sidebar-bg); border: 1px solid var(--border-color); color: var(--text-color); border-radius: 0 4px 4px 0; }
.encryption-options { margin-top: 1.5rem; padding-top: 1.5rem; border-top: 1px solid var(--border-color); }
input[type="checkbox"].toggle { -webkit-appearance: none; appearance: none; width: 40px; height: 22px; background-color: #4a5568; border-radius: 11px; position: relative; cursor: pointer; transition: background-color 0.2s; }
input[type="checkbox"].toggle::before { content: ''; position: absolute; width: 18px; height: 18px; border-radius: 50%; background-color: white; top: 2px; left: 2px; transition: transform 0.2s; }
input[type="checkbox"].toggle:checked { background-color: var(--primary-color); }
input[type="checkbox"].toggle:checked::before { transform: translateX(18px); }
</style>