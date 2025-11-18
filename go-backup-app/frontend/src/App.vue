<template>
  <div id="app-container">
    <!-- Main View: Home Screen -->
    <main class="main-content" v-if="currentScreen === 'home'">
      <div class="home-screen">
        <!-- <div class="title-badge">BETA v1.0</div> -->
        <h1 class="home-title">CcBackup</h1>
        <div class="home-actions">
          <div class="action-card color-cyan" @click="navigateTo('backup')">
            <div class="card-decoration"></div>
            <div class="icon">
              <!-- Icon preserved but styled via CSS filter/container -->
              <svg height="4rem" width="4rem" version="1.1" viewBox="0 0 511.998 511.998" xml:space="preserve">
                  <g> <path style="fill:#000;" d="M467.331,202.728c-8.994,0-16.284,7.29-16.284,16.284v260.42H105.826v-156.66h62.825v48.098 c0,6.279,3.61,12,9.279,14.7c2.231,1.063,4.623,1.584,7.002,1.584c3.667,0,7.302-1.238,10.259-3.636l141.64-114.871 c3.812-3.091,6.027-7.738,6.027-12.648c0-4.91-2.215-9.555-6.027-12.648l-141.64-114.871c-4.877-3.955-11.593-4.753-17.263-2.052 c-5.668,2.702-9.279,8.42-9.279,14.7v51.368H44.668c-8.994,0-16.284,7.29-16.284,16.284v97.706c0,8.994,7.29,16.284,16.284,16.284 h28.59v172.944c0,8.994,7.29,16.284,16.284,16.284h377.789c8.994,0,16.284-7.29,16.284-16.284V219.012 C483.615,210.018,476.324,202.728,467.331,202.728z M60.952,225.066h123.983c8.994,0,16.284-7.29,16.284-16.284v-33.48 l99.503,80.697l-99.503,80.697v-30.208c0-8.994-7.29-16.284-16.284-16.284H60.952V225.066z"></path>
                  <path style="fill:#000;" d="M89.542,138.318c8.994,0,16.284-7.29,16.284-16.284V32.568h225.476v89.466 c0,8.994,7.29,16.284,16.284,16.284H464.78c6.689,0,12.698-4.092,15.151-10.316c2.452-6.224,0.848-13.315-4.045-17.878 L362.488,4.376c-0.124-0.116-0.259-0.217-0.386-0.327c-0.192-0.168-0.383-0.335-0.583-0.493c-0.2-0.16-0.405-0.309-0.612-0.458 c-0.217-0.156-0.432-0.311-0.656-0.458c-0.199-0.129-0.402-0.248-0.606-0.368c-0.246-0.147-0.493-0.29-0.747-0.423 c-0.191-0.099-0.386-0.187-0.58-0.28c-0.275-0.13-0.55-0.261-0.835-0.375c-0.189-0.077-0.383-0.14-0.575-0.21 c-0.295-0.106-0.588-0.213-0.889-0.303c-0.207-0.062-0.417-0.107-0.625-0.161c-0.291-0.073-0.58-0.153-0.878-0.21 c-0.267-0.052-0.537-0.085-0.806-0.124c-0.246-0.034-0.489-0.081-0.738-0.104C352.445,0.029,351.914,0,351.38,0h-3.794H89.542 c-8.994,0-16.284,7.29-16.284,16.284v105.75C73.258,131.028,80.548,138.318,89.542,138.318z M363.87,105.75V50.197l59.572,55.553 H363.87z"></path> </g>
              </svg>
            </div>
            <h2>备份数据</h2>
            <p>创建新的数据备份</p>
          </div>
          <div class="action-card color-yellow" @click="navigateTo('restore')">
             <div class="card-decoration"></div>
            <div class="icon">
               <svg height="4rem" width="4rem" version="1.1" viewBox="0 0 512.002 512.002" xml:space="preserve">
                  <g> <path style="fill:#000;" d="M467.334,202.728c-8.994,0-16.284,7.29-16.284,16.284v260.42H105.829V326.566l70.225,56.953 c2.957,2.399,6.592,3.636,10.259,3.636c2.379,0,4.773-0.521,7.002-1.584c5.668-2.702,9.279-8.42,9.279-14.7v-51.368h123.983 c8.994,0,16.284-7.29,16.284-16.284v-97.706c0-8.994-7.29-16.284-16.284-16.284H202.593v-48.1c0-6.279-3.61-12-9.279-14.7 c-5.665-2.702-12.384-1.902-17.263,2.052L34.412,243.353c-3.812,3.091-6.027,7.738-6.027,12.648c0,4.91,2.215,9.555,6.027,12.648 l38.847,31.505v195.564c0,8.994,7.29,16.284,16.284,16.284h377.789c8.994,0,16.284-7.29,16.284-16.284V219.012 C483.618,210.018,476.327,202.728,467.334,202.728z M170.027,175.302v30.208c0,8.994,7.29,16.284,16.284,16.284h123.983v65.138 H186.311c-8.994,0-16.284,7.29-16.284,16.284v33.48l-99.503-80.697L170.027,175.302z"></path>
                  <path style="fill:#000;" d="M89.545,138.318c8.994,0,16.284-7.29,16.284-16.284V32.568h225.476v89.466 c0,8.994,7.29,16.284,16.284,16.284h117.194c6.689,0,12.698-4.092,15.151-10.316c2.452-6.224,0.848-13.315-4.045-17.878 L362.491,4.376c-0.124-0.116-0.259-0.217-0.386-0.327c-0.192-0.168-0.383-0.335-0.583-0.493c-0.2-0.16-0.405-0.309-0.612-0.458 c-0.217-0.156-0.432-0.311-0.656-0.458c-0.199-0.129-0.402-0.248-0.606-0.368c-0.246-0.147-0.493-0.29-0.747-0.423 c-0.191-0.099-0.386-0.187-0.58-0.28c-0.275-0.13-0.55-0.261-0.835-0.375c-0.189-0.077-0.383-0.14-0.575-0.21 c-0.295-0.106-0.588-0.213-0.889-0.303c-0.207-0.062-0.417-0.107-0.625-0.161c-0.292-0.073-0.58-0.153-0.878-0.21 c-0.267-0.052-0.537-0.085-0.806-0.124c-0.246-0.034-0.489-0.081-0.738-0.104C352.448,0.029,351.918,0,351.384,0h-3.794H89.545 c-8.994,0-16.284,7.29-16.284,16.284v105.75C73.261,131.028,80.551,138.318,89.545,138.318z M363.873,105.75V50.197l59.572,55.553 H363.873z"></path> </g> </svg>
            </div>
            <h2>恢复数据</h2>
            <p>从备份文件中恢复数据</p>
          </div>
        </div>
      </div>
    </main>

    <!-- Nested Views: Backup/Restore pages -->
    <main class="main-content" v-else>
      <div class="view-header">
        <button class="back-button" @click="navigateBack">← BACK</button>
        <h2 class="view-title">{{ viewTitle.toUpperCase() }}</h2>
      </div>

      <!-- Backup View -->
      <div v-if="currentScreen === 'backup'" class="view">
        <div class="stepper">
          <div class="step" :class="{ active: backupStep === 1, completed: backupStep > 1 }">
            <div class="step-box">1</div>
            <div class="step-title">选择项目</div>
          </div>
          <div class="step-line" :class="{ completed: backupStep > 1 }"></div>
          <div class="step" :class="{ active: backupStep === 2, completed: backupStep > 2 }">
            <div class="step-box">2</div>
            <div class="step-title">高级筛选</div>
          </div>
          <div class="step-line" :class="{ completed: backupStep > 2 }"></div>
          <div class="step" :class="{ active: backupStep === 3 }">
            <div class="step-box">3</div>
            <div class="step-title">压缩加密</div>
          </div>
        </div>

        <!-- Step 1: Select Paths & Files -->
        <div v-if="backupStep === 1">
          <div class="card neo-card">
            <div class="card-label">Step 1</div>
            <h3>选择要备份的文件或文件夹</h3>

            <!-- Profile Section -->
            <div class="profile-section">
              <h4 class="profile-title">档案快速选择</h4>
              <p class="description">选择一个预设或自定义的档案来快速添加常用文件夹。</p>
              <div class="profile-actions">
                <div class="profile-buttons">
                  <button v-for="profile in profiles" :key="profile.id" @click="applyProfile(profile)"
                          class="profile-btn">
                    {{ profile.name }}
                  </button>
                </div>
                <button class="secondary" @click="openSaveProfileModal" :disabled="selectedBackupFileCount === 0">
                  保存为档案
                </button>
              </div>
            </div>
            <hr class="card-divider">

            <p class="description">或者手动选择，它们将显示在下面的列表中。</p>
            <div class="action-bar-left">
              <button class="brutal-btn" @click="selectBackupSources('files')">选择文件</button>
              <button class="brutal-btn" @click="selectBackupSources('dirs')">选择文件夹</button>
            </div>
          </div>

          <div class="card file-list-card neo-card" v-if="backupFiles.length > 0">
            <div class="file-list-header">
              <!-- Breadcrumbs -->
              <div class="breadcrumbs">
                <span v-for="(part, index) in pathStack" :key="part.path" @click="navigateToBreadcrumb(index)"
                      class="breadcrumb-item">
                  {{ part.name }}
                </span>
              </div>
              <span class="badge-count">{{ selectedBackupFileCount }} ITEMS</span>
            </div>

            <!-- Loading indicator -->
            <div v-if="isBrowsing" class="loading-overlay">
              <p>LOADING...</p>
            </div>

            <div class="table-container">
              <table class="file-table" :class="{ 'is-loading': isBrowsing }">
                <thead>
                <tr>
                  <th class="col-checkbox"><input type="checkbox" @change="toggleSelectAllCurrentView"
                                                  :checked="allCurrentViewFilesSelected"
                                                  :indeterminate="isCurrentViewIndeterminate"></th>
                  <th @click="sortCurrentViewItems('name')" class="sortable">名称 <span
                      v-if="sort.key === 'name'">{{ sort.order === 'asc' ? '▲' : '▼' }}</span></th>
                  <th @click="sortCurrentViewItems('size')" class="sortable">大小 <span
                      v-if="sort.key === 'size'">{{ sort.order === 'asc' ? '▲' : '▼' }}</span></th>
                  <th @click="sortCurrentViewItems('modTime')" class="sortable">修改时间 <span
                      v-if="sort.key === 'modTime'">{{ sort.order === 'asc' ? '▲' : '▼' }}</span></th>
                </tr>
                </thead>
                <tbody>
                <tr v-for="file in sortedCurrentViewItems" :key="file.path" :class="{'row-selected': file.selected}">
                  <td><input type="checkbox" v-model="file.selected" @change="updateSelection(file)"></td>
                  <td class="col-name">
                    <span class="file-icon">{{ file.isDir ? '📁' : '📄' }}</span>
                    <span :class="{ 'dir-link': file.isDir }" @click="file.isDir ? enterDirectory(file) : null">
                       {{ file.name }}
                    </span>
                  </td>
                  <td class="font-mono">{{ formatSize(file.size) }}</td>
                  <td class="font-mono">{{ formatDate(file.modTime) }}</td>
                </tr>
                <tr v-if="!isBrowsing && sortedCurrentViewItems.length === 0">
                  <td colspan="5" class="empty-dir-msg">EMPTY FOLDER</td>
                </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="action-bar">
            <button class="primary brutal-btn-lg" @click="backupStep = 2" :disabled="selectedBackupFileCount === 0">
              下一步 ({{ selectedBackupFileCount }})
            </button>
          </div>
        </div>

        <!-- Step 2: Advanced Filters -->
        <div v-if="backupStep === 2">
          <div class="card neo-card">
            <div class="card-label">Step 2</div>
            <h3>高级筛选 (可选)</h3>
            <p class="description">
              定义过滤规则。如果不需要，直接点击下一步。
            </p>
            <div class="filter-grid">
              <div class="filter-group">
                <label>包含名称 (Match)</label>
                <textarea v-model="filters.includeNames" placeholder="*.log"></textarea>
              </div>
              <div class="filter-group">
                <label>排除名称 (Ignore)</label>
                <textarea v-model="filters.excludeNames" placeholder="*.tmp"></textarea>
              </div>
              <div class="filter-group">
                <label>最小大小</label>
                <div class="size-input">
                  <input type="number" v-model.number="filters.minSizeValue" min="0">
                  <select v-model="filters.minSizeUnit">
                    <option>Bytes</option>
                    <option>KB</option>
                    <option>MB</option>
                    <option>GB</option>
                  </select>
                </div>
              </div>
              <div class="filter-group">
                <label>最大大小</label>
                <div class="size-input">
                  <input type="number" v-model.number="filters.maxSizeValue" min="0">
                  <select v-model="filters.maxSizeUnit">
                    <option>Bytes</option>
                    <option>KB</option>
                    <option>MB</option>
                    <option>GB</option>
                  </select>
                </div>
              </div>
            </div>
          </div>
          <div class="action-bar">
            <button @click="backupStep = 1" class="brutal-btn">上一步</button>
            <button class="primary brutal-btn-lg" @click="backupStep = 3">下一步</button>
          </div>
        </div>

        <!-- Step 3: Encryption & Start -->
        <div v-if="backupStep === 3">
          <div v-if="!inProgress">
            <div class="card neo-card">
              <div class="card-label">Step 3</div>
              <h3>压缩与加密</h3>
              <!-- Compression Switch -->
              <div class="input-group switch-group">
                <label>启用压缩</label>
                <input type="checkbox" v-model="compression.enabled" class="toggle"/>
              </div>
              
              <hr class="card-divider">
              <!-- Encryption Section -->
              <div class="input-group switch-group">
                <label>启用加密</label>
                <input type="checkbox" v-model="encryption.enabled" class="toggle"/>
              </div>
              <div v-if="encryption.enabled" class="encryption-options">
                <div class="input-group">
                  <label>密码:</label>
                  <input type="password" v-model="encryption.password" placeholder="PASSWORD"/>
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
            <div class="card neo-card">
              <h3>目标位置</h3>
              <div class="input-group">
                <input v-model="backupDest" readonly type="text" placeholder="选择目录..."/>
                <button class="brutal-btn" @click="selectDestDir">浏览</button>
              </div>
            </div>
            <div class="action-bar">
              <button @click="backupStep = 2" class="brutal-btn">上一步</button>
              <button class="primary brutal-btn-lg" @click="doBackup"
                      :disabled="!backupDest || (encryption.enabled && !encryption.password)">
                开始备份
              </button>
            </div>
          </div>
          <!-- Backup Progress View -->
          <div v-if="currentScreen === 'backup' && backupStep === 3 && inProgress" class="progress-view">
            <h3>正在处理</h3>
            <div class="progress-container">
                <progress :value="progress.value" :max="100"></progress>
            </div>
            <div class="progress-stage">{{ getProgressStageText() }}</div>
            <p class="status-text">> {{ statusMessage }}</p>
            <div class="action-bar center">
              <button class="danger brutal-btn" @click="stopOperation">停止备份</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Restore View -->
      <div v-if="currentScreen === 'restore'" class="view">
        <div v-if="!inProgress">
          <div class="card neo-card">
             <div class="card-label">HISTORY</div>
            <h3>从历史记录恢复</h3>
            
            <div class="backup-history-list">
              <div v-for="item in backupHistory" :key="item.ID"
                   class="history-item-wrapper"
                   :class="{ 'expanded': expandedBackupId === item.ID }">
                <div class="history-item" @click="toggleBackupDetails(item)">
                  <div class="history-item-main">
                    <span class="history-file">{{ item.FileName }}</span>
                    <span class="history-date font-mono">{{ formatDate(item.CreatedAt) }}</span>
                  </div>
                  <small class="history-path font-mono">{{ item.BackupPath }}</small>
                </div>
                <!-- Expanded Details View -->
                <div v-if="expandedBackupId === item.ID" class="backup-details">
                  <h4>包含内容:</h4>
                  <ul>
                    <li v-for="(path, index) in item.SourcePaths.split('\n').filter(p => p)" :key="index">
                      {{ path }}
                    </li>
                  </ul>
                  <button class="primary brutal-btn small-btn" @click.stop="restoreFile = item.BackupPath; expandedBackupId = item.ID">选择此项</button>
                </div>
              </div>
              <div v-if="!backupHistory.length" class="history-empty">
                NO RECORDS FOUND
              </div>
            </div>
          </div>
          <div class="card neo-card">
             <div class="card-label">MANUAL</div>
            <h3>手动选择</h3>
            <div class="input-group">
              <label>源文件:</label>
              <input v-model="restoreFile" readonly type="text" placeholder=".qbak FILE"/>
              <button class="brutal-btn" @click="selectRestoreFileManually">选择</button>
            </div>
            <div class="input-group">
              <label>恢复至:</label>
              <input v-model="restoreDir" readonly type="text" placeholder="TARGET DIR"/>
              <button class="brutal-btn" @click="selectRestoreDir">选择</button>
            </div>
          </div>
          <div class="action-bar">
            <button class="primary brutal-btn-lg" @click="doRestore()" :disabled="!restoreFile || !restoreDir || inProgress">
              开始恢复
            </button>
          </div>
        </div>
        <!-- Restore Progress View -->
        <div v-if="currentScreen === 'restore' && inProgress" class="progress-view">
          <h3>正在恢复</h3>
          <div class="progress-container">
              <progress :value="progress.value" :max="100"></progress>
          </div>
          <div class="progress-stage">{{ getProgressStageText() }}</div>
          <div class="log-box neo-log">
            <p v-for="(msg, index) in logMessages" :key="index">> {{ msg }}</p>
          </div>
          <div class="action-bar center">
            <button class="danger brutal-btn" @click="stopOperation">停止恢复</button>
          </div>
        </div>
      </div>
    </main>

    <!-- Log Section (Global) -->
    <!-- <div class="log-card-container">
      <div class="log-card neo-card small">
        <h3>SYSTEM LOG</h3>
        <p class="status">> {{ statusMessage }}</p>
      </div>
    </div> -->

    <!-- Password Modal -->
    <div v-if="isPasswordModalVisible" class="modal-overlay">
      <div class="modal-content neo-modal">
        <div class="modal-header-strip">SECURITY</div>
        <h3>需要密码</h3>
        <p>此备份文件已加密。</p>
        <div class="input-group modal-input">
          <input
              type="password"
              v-model="restorePasswordInput"
              @keyup.enter="submitPasswordAndRetryRestore"
              placeholder="ENTER PASSWORD"
              ref="passwordInputRef"
          />
        </div>
        <div class="modal-actions">
          <button class="brutal-btn" @click="cancelPasswordPrompt">取消</button>
          <button class="primary brutal-btn" @click="submitPasswordAndRetryRestore">解锁</button>
        </div>
      </div>
    </div>

    <!-- Success Modal -->
    <div v-if="showSuccessModal" class="modal-overlay">
      <div class="modal-content neo-modal success-theme">
         <div class="modal-header-strip success-strip">SUCCESS</div>
        <h3>操作成功</h3>
        <p>{{ successMessage }}</p>
        <div class="modal-actions">
          <button class="primary brutal-btn" @click="closeSuccessModal">确定</button>
        </div>
      </div>
    </div>

    <!-- Error Modal -->
    <div v-if="showErrorModal" class="modal-overlay">
      <div class="modal-content neo-modal error-theme">
         <div class="modal-header-strip error-strip">ERROR</div>
        <h3>操作失败</h3>
        <p>{{ successMessage }}</p>
        <div class="modal-actions">
          <button class="primary brutal-btn" @click="closeErrModal">确定</button>
        </div>
      </div>
    </div>

    <!-- Conflict Modal -->
    <div v-if="isConflictModalVisible" class="modal-overlay">
      <div class="modal-content neo-modal warning-theme">
        <div class="modal-header-strip warning-strip">CONFLICT</div>
        <h3>文件冲突</h3>
        <p>目标位置已存在文件：</p>
        <p class="font-mono"><strong>{{ conflictInfo.path }}</strong></p>
        <div class="modal-actions">
          <button class="brutal-btn" @click="resolveConflict('skip')">跳过</button>
          <button class="brutal-btn" @click="resolveConflict('keep_both')">保留两者</button>
          <button class="primary brutal-btn" @click="resolveConflict('overwrite')">覆盖</button>
        </div>
      </div>
    </div>

    <!-- Profile Save Modal -->
    <div v-if="isProfileModalVisible" class="modal-overlay">
      <div class="modal-content neo-modal">
        <div class="modal-header-strip">PROFILE</div>
        <h3>保存新档案</h3>
        <p>为选中的 {{ selectedBackupFileCount }} 个项目命名。</p>
        <div class="input-group modal-input">
          <input
              type="text"
              v-model="newProfileName"
              @keyup.enter="confirmSaveProfile"
              placeholder="PROFILE NAME"
              ref="profileNameInputRef"
          />
        </div>
        <div class="modal-actions">
          <button class="brutal-btn" @click="closeSaveProfileModal">取消</button>
          <button class="primary brutal-btn" @click="confirmSaveProfile" :disabled="!newProfileName.trim()">保存</button>
        </div>
      </div>
    </div>

  </div>

</template>

<script setup>
// Script逻辑保持不变
import {computed, nextTick, onMounted, reactive, ref} from 'vue';
import {
  CreateProfile,
  GetBackupHistory,
  GetFileMetadata,
  GetProfiles,
  ListDirectory,
  OpenInExplorer,
  ResolveConflict,
  SelectDirectory,
  SelectFiles,
  StartBackup,
  StartRestore,
  StopOperation
} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

onMounted(() => {
  EventsOn("log_message", (data) => {
    logMessages.value.unshift(data);
    statusMessage.value = data;
    if (logMessages.value.length > 200) logMessages.value.pop();
  });
  EventsOn("progress_update", (p) => {
    statusMessage.value = p.message;
  });
  EventsOn("conflict_detected", (data) => {
    conflictInfo.path = data.path;
    conflictInfo.requestID = data.requestID;
    isConflictModalVisible.value = true;
  });

  fetchBackupHistory();
  fetchProfiles();
});

const currentScreen = ref('home');
const inProgress = ref(false);
const statusMessage = ref('READY');
const logMessages = ref([]);
const progress = reactive({value: 0, max: 100});
const isPasswordModalVisible = ref(false);
const restorePasswordInput = ref('');
const passwordInputRef = ref(null);

const progressStage = ref(''); 
const progressStartTime = ref(0);

const viewTitle = computed(() => {
  if (currentScreen.value === 'backup') return 'Create Backup';
  if (currentScreen.value === 'restore') return 'Restore Data';
  return 'GoBackup';
});

function navigateTo(screen) {
  currentScreen.value = screen;
  if (screen === 'backup') resetBackupState();
  if (screen === 'restore') resetRestoreState();
}

function navigateBack() {
  currentScreen.value = 'home';
  if (inProgress.value) {
    stopOperation();
  }
}

const backupStep = ref(1);
const backupFiles = ref([]);
const backupDest = ref('');
const sort = reactive({key: 'name', order: 'asc'});
const filters = ref({
  includeNames: '', excludeNames: '', minSizeValue: 0, minSizeUnit: 'Bytes',
  maxSizeValue: 0, maxSizeUnit: 'Bytes', newerThan: null, olderThan: null,
});
const compression = reactive({enabled: true});
const encryption = reactive({enabled: false, password: '', algorithm: 'AES-256'});

const pathStack = ref([{name: 'ROOT', path: 'root'}]);
const currentViewItems = ref([]);
const isBrowsing = ref(false);
const fileSelectionMap = reactive(new Map());

const profiles = ref([]);
const isProfileModalVisible = ref(false);
const newProfileName = ref('');
const profileNameInputRef = ref(null);

async function fetchProfiles() {
  try {
    profiles.value = await GetProfiles();
  } catch (e) {
    statusMessage.value = `Failed to load profiles: ${e}`;
  }
}

async function applyProfile(profile) {
  backupFiles.value = [];
  fileSelectionMap.clear();
  pathStack.value = [{name: 'ROOT', path: 'root'}];
  currentViewItems.value = [];

  const paths = profile.paths.split('\n').filter(p => p.trim() !== '');
  if (paths.length > 0) {
    try {
      statusMessage.value = `Loading profile '${profile.name}'...`;
      const metadata = await GetFileMetadata(paths);
      metadata.forEach(m => {
        if (!backupFiles.value.some(f => f.path === m.path)) {
          backupFiles.value.push({...m, selected: true});
        }
        fileSelectionMap.set(m.path, true);
      });
      await loadDirectoryView('root');
    } catch (error) {
      statusMessage.value = `Error: ${error}`;
    }
  }
}

function openSaveProfileModal() {
  isProfileModalVisible.value = true;
  nextTick(() => {
    profileNameInputRef.value?.focus();
  });
}

function closeSaveProfileModal() {
  isProfileModalVisible.value = false;
  newProfileName.value = '';
}

async function confirmSaveProfile() {
  const name = newProfileName.value.trim();
  if (!name) {
    statusMessage.value = "Name cannot be empty.";
    return;
  }
  const selectedPaths = Array.from(fileSelectionMap.entries())
      .filter(([, selected]) => selected)
      .map(([path]) => path);

  if (selectedPaths.length === 0) {
    statusMessage.value = "No items selected.";
    return;
  }

  try {
    await CreateProfile(name, selectedPaths);
    statusMessage.value = `Profile '${name}' saved.`;
    await fetchProfiles();
    closeSaveProfileModal();
  } catch (error) {
    statusMessage.value = `Error: ${error}`;
    successMessage.value = `Error: ${error}`;
    showErrorModal.value = true;
  }
}

function resetBackupState() {
  backupStep.value = 1;
  backupFiles.value = [];
  backupDest.value = '';
  inProgress.value = false;
  isProfileModalVisible.value = false;
  newProfileName.value = '';
  compression.enabled = true;
  encryption.enabled = false;
  encryption.password = '';
  pathStack.value = [{name: 'ROOT', path: 'root'}];
  currentViewItems.value = [];
  fileSelectionMap.clear();
}

function getProgressStageText() {
  switch (progressStage.value) {
    case 'compressing': return 'COMPRESSING...';
    case 'encrypting': return 'ENCRYPTING...';
    case 'archiving': return 'ARCHIVING...';
    case 'decrypting': return 'DECRYPTING...';
    case 'decompressing': return 'DECOMPRESSING...';
    case 'restoring': return 'RESTORING...';
    default: return 'PROCESSING...';
  }
}

async function selectBackupSources(type) {
  try {
    const paths = await SelectFiles(type === 'dirs');
    if (paths && paths.length > 0) {
      const metadata = await GetFileMetadata(paths);
      metadata.forEach(m => {
        if (!backupFiles.value.some(f => f.path === m.path)) {
          backupFiles.value.push({...m, selected: true});
        }
        fileSelectionMap.set(m.path, true);
      });
      await loadDirectoryView('root');
    }
  } catch (error) {
    statusMessage.value = `Error: ${error}`;
  }
}

async function loadDirectoryView(path) {
  isBrowsing.value = true;
  statusMessage.value = `Loading ${path}...`;
  try {
    let items;
    if (path === 'root') {
      items = [...backupFiles.value];
    } else {
      items = await ListDirectory(path);
    }
    currentViewItems.value = items.map(item => ({
      ...item,
      selected: fileSelectionMap.get(item.path) || false
    }));
  } catch (error) {
    statusMessage.value = `Error: ${error}`;
    currentViewItems.value = [];
  } finally {
    isBrowsing.value = false;
    statusMessage.value = 'READY';
  }
}

async function enterDirectory(dir) {
  pathStack.value.push({name: dir.name, path: dir.path});
  await loadDirectoryView(dir.path);
}

async function navigateToBreadcrumb(index) {
  pathStack.value = pathStack.value.slice(0, index + 1);
  const targetPath = pathStack.value[pathStack.value.length - 1].path;
  await loadDirectoryView(targetPath);
}

function updateSelection(file) {
  fileSelectionMap.set(file.path, file.selected);
}

const sortedCurrentViewItems = computed(() => {
  return [...currentViewItems.value].sort((a, b) => {
    if (a.isDir !== b.isDir) {
      return a.isDir ? -1 : 1;
    }
    let modifier = sort.order === 'asc' ? 1 : -1;
    let valA = a[sort.key];
    let valB = b[sort.key];
    if (typeof valA === 'string') {
      return valA.localeCompare(valB) * modifier;
    }
    if (valA < valB) return -1 * modifier;
    if (valA > valB) return 1 * modifier;
    return 0;
  });
});

function sortCurrentViewItems(key) {
  if (sort.key === key) {
    sort.order = sort.order === 'asc' ? 'desc' : 'asc';
  } else {
    sort.key = key;
    sort.order = 'asc';
  }
}

const allCurrentViewFilesSelected = computed(() => {
  if (currentViewItems.value.length === 0) return false;
  return currentViewItems.value.every(f => f.selected);
});

const isCurrentViewIndeterminate = computed(() => {
  if (currentViewItems.value.length === 0) return false;
  const selectedCount = currentViewItems.value.filter(f => f.selected).length;
  return selectedCount > 0 && selectedCount < currentViewItems.value.length;
});

function toggleSelectAllCurrentView(event) {
  const isChecked = event.target.checked;
  currentViewItems.value.forEach(f => {
    f.selected = isChecked;
    fileSelectionMap.set(f.path, isChecked);
  });
}

const selectedBackupFileCount = computed(() => {
  let count = 0;
  for (const selected of fileSelectionMap.values()) {
    if (selected) count++;
  }
  return count;
});

const selectDestDir = async () => {
  if (inProgress.value) return;
  const dir = await SelectDirectory();
  if (dir) backupDest.value = dir;
};

function calculateLogProgress(elapsedTime) {
  const maxTime = 200 * 1000;
  if (elapsedTime >= maxTime) return 90;
  const progress = Math.log10(1 + 9 * (elapsedTime / maxTime)) * 100;
  return Math.min(99, Math.max(0, progress));
}

function updateProgress() {
  if (!inProgress.value || progressStartTime.value === 0) return;

  if (isConflictModalVisible.value) {
    progress.value = progress.value;
  } else {
    const elapsed = Date.now() - progressStartTime.value;
    progress.value = calculateLogProgress(elapsed);
  }

  if (currentScreen.value === 'backup' && encryption.enabled) {
    if (progress.value >= 44 && progressStage.value === 'compressing') {
      progressStage.value = 'encrypting';
    } else if (progress.value >= 88 && progressStage.value === 'encrypting') {
      progressStage.value = 'archiving';
    }
  } else if (currentScreen.value === 'backup' && !encryption.enabled) {
    if (progress.value >= 70 && progressStage.value === 'archiving') {
      progressStage.value = 'archiving';
    }
  }

  if (currentScreen.value === 'restore') {
    if (progress.value >= 44 && progressStage.value === 'decrypting') {
      progressStage.value = 'decompressing';
    } else if (progress.value >= 88 && progressStage.value === 'decompressing') {
      progressStage.value = 'restoring';
    }
  }
  
  if (inProgress.value) {
    setTimeout(updateProgress, 100);
  }
}

async function doBackup() {
  const selectedPaths = Array.from(fileSelectionMap.entries())
      .filter(([, selected]) => selected)
      .map(([path]) => path);
  if (selectedPaths.length === 0 || !backupDest.value) {
    statusMessage.value = "Please select files and destination.";
    return;
  }
  if (encryption.enabled && !encryption.password) {
    statusMessage.value = "Password required for encryption.";
    return;
  }

  inProgress.value = true;
  statusMessage.value = "Starting backup...";
  logMessages.value = [];
  progress.value = 0;
  progressStartTime.value = Date.now();
  progressStage.value = 'compressing';

  updateProgress();

  const multipliers = {Bytes: 1, KB: 1024, MB: 1024 ** 2, GB: 1024 ** 3};
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
      UseCompression: compression.enabled,
      useEncryption: encryption.enabled,
      encryptionAlgorithm: encryption.algorithm,
      encryptionPassword: encryption.password,
    });

    statusMessage.value = result;
    progress.value = 100;

    if (result.includes("成功")) {
      successMessage.value = "BACKUP COMPLETED SUCCESSFULLY";
      showSuccessModal.value = true;
    }
  } catch (error) {
    successMessage.value = `${error}`;
    showErrorModal.value = true;
    statusMessage.value = `Error: ${error}`;
  } finally {
    inProgress.value = false;
    await fetchBackupHistory();
  }
}

const backupHistory = ref([]);
const restoreFile = ref('');
const restoreDir = ref('');
const expandedBackupId = ref(null);

function resetRestoreState() {
  restoreFile.value = '';
  restoreDir.value = '';
  backupHistory.value = [];
  restorePasswordInput.value = '';
  inProgress.value = false;
  isPasswordModalVisible.value = false;
  expandedBackupId.value = null;
  fetchBackupHistory();
}

function toggleBackupDetails(item) {
  restoreFile.value = item.BackupPath;
  if (expandedBackupId.value === item.ID) {
    expandedBackupId.value = null;
  } else {
    expandedBackupId.value = item.ID;
  }
}

async function fetchBackupHistory() {
  try {
    backupHistory.value = await GetBackupHistory();
  } catch (e) {
    statusMessage.value = `Cannot load history: ${e}`;
  }
}

async function selectRestoreFileManually() {
  const path = await SelectFiles(false);
  if (path && path.length > 0) {
    restoreFile.value = path[0];
  }
}

const selectRestoreDir = async () => {
  if (inProgress.value) return;
  const dir = await SelectDirectory();
  if (dir) restoreDir.value = dir;
};

async function doRestore(password) {
  let cleanPassword = (typeof password === 'string') ? password : '';
  if (!restoreFile.value || !restoreDir.value) {
    statusMessage.value = "Select backup file and destination.";
    return;
  }

  inProgress.value = true;
  statusMessage.value = "Restoring...";
  if (cleanPassword === '') {
    logMessages.value = [];
  }

  progress.value = 0;
  progressStartTime.value = Date.now();
  progressStage.value = 'decrypting';

  updateProgress();

  try {
    const result = await StartRestore({
      backupFile: restoreFile.value,
      restoreDir: restoreDir.value,
      password: cleanPassword,
    });
    statusMessage.value = result;

    progress.value = 100;

    if (result.includes("成功")) {
      successMessage.value = "RESTORE COMPLETED SUCCESSFULLY";
      showSuccessModal.value = true;
      await OpenInExplorer(restoreDir.value);
    }
  } catch (error) {
    if (typeof error === 'string' && error.includes("password_required")) {
      inProgress.value = false;
      statusMessage.value = "Password required.";
      isPasswordModalVisible.value = true;
      await nextTick();
      passwordInputRef.value?.focus();
    } else {
      statusMessage.value = `Error: ${error}`;
      showErrorModal.value = true;
      successMessage.value = `${error}`;
      if (error.includes("unexpected EOF")) {
        statusMessage.value = "Incorrect password.";
        successMessage.value = "密码错误";
      }
    }
  } finally {
    if (!isPasswordModalVisible.value) {
      inProgress.value = false;
    }
  }
}

function submitPasswordAndRetryRestore() {
  if (!restorePasswordInput.value) {
    successMessage.value = `Password cannot be empty`;
    showErrorModal.value = true;
    return;
  }
  isPasswordModalVisible.value = false;
  inProgress.value = true;
  doRestore(restorePasswordInput.value);
  restorePasswordInput.value = '';
}

function cancelPasswordPrompt() {
  isPasswordModalVisible.value = false;
  statusMessage.value = "Cancelled.";
  inProgress.value = false;
  restorePasswordInput.value = '';
}

function stopOperation() {
  StopOperation();
  inProgress.value = false;
  statusMessage.value = "STOPPED.";
  progressStartTime.value = 0;
  progress.value = 0;
  progressStage.value = '';

  if (currentScreen.value === 'backup') {
    backupStep.value = Math.max(1, backupStep.value - 1);
  }
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const formatDate = (dateString) => {
  if (!dateString) return '-';
  return new Date(dateString).toLocaleString('zh-CN');
};

const showSuccessModal = ref(false);
const successMessage = ref('');
const showErrorModal = ref(false);

function closeSuccessModal() {
  showSuccessModal.value = false;
  currentScreen.value = 'home';
  resetBackupState();
  resetRestoreState();
}

function closeErrModal() {
  showErrorModal.value = false;
  successMessage.value = '';
}

const isConflictModalVisible = ref(false);
const conflictInfo = reactive({
  path: '',
  requestID: '',
});

async function resolveConflict(resolution) {
  if (!conflictInfo.requestID) return;

  try {
    await ResolveConflict(conflictInfo.requestID, resolution);
    isConflictModalVisible.value = false;
    conflictInfo.path = '';
    conflictInfo.requestID = '';
  } catch (error) {
    statusMessage.value = `Error: ${error}`;
    showErrorModal.value = true;
    successMessage.value = `Error: ${error}`;
    isConflictModalVisible.value = false;
  }
}

</script>

<style>
/* NEO-BRUTALISM CSS STYLES */
:root {
  --bg-color: #fdf6e3; /* Warm off-white / Beige */
  --card-bg: #ffffff;
  --text-color: #000000;
  --text-color-muted: #555555;
  
  /* Accent Colors */
  --primary-color: #8b5cf6; /* Violet */
  --secondary-color: #ffeb3b; /* Yellow */
  --cyan-color: #67e8f9;
  --danger-color: #ff6b6b; /* Salmon/Red */
  --success-color: #4ade80; /* Green */

  --border-width: 3px;
  --border-color: #000000;
  --shadow-distance: 4px;
  
  --font-main: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  --font-mono: "Courier New", Courier, monospace;
}

body, html {
  margin: 0;
  font-family: var(--font-main);
  background-color: var(--bg-color);
  color: var(--text-color);
  font-size: 16px;
  overflow: hidden;
}

#app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

/* Common Classes */
.font-mono { font-family: var(--font-mono); }
.center { justify-content: center !important; }

/* Layout & Main Content */
.main-content {
  flex-grow: 1;
  padding: 2rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  position: relative;
}

/* Neo-Brutal Card Style */
.neo-card, .card {
  background-color: var(--card-bg);
  border: var(--border-width) solid var(--border-color);
  box-shadow: var(--shadow-distance) var(--shadow-distance) 0px var(--border-color);
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  border-radius: 0px; /* Sharp corners */
  position: relative;
}

.card-label {
  position: absolute;
  top: -12px;
  left: 20px;
  background: var(--primary-color);
  color: white;
  padding: 0 10px;
  border: var(--border-width) solid var(--border-color);
  font-weight: bold;
  font-size: 0.8rem;
  text-transform: uppercase;
}

.card h3 {
  margin-top: 0;
  font-size: 1.5rem;
  font-weight: 900;
  text-transform: uppercase;
  border-bottom: var(--border-width) solid var(--border-color);
  padding-bottom: 0.5rem;
  margin-bottom: 1rem;
}

/* Buttons */
button, .brutal-btn {
  padding: 0.6rem 1.2rem;
  background-color: white;
  border: var(--border-width) solid var(--border-color);
  color: var(--text-color);
  border-radius: 0px;
  cursor: pointer;
  font-weight: 700;
  text-transform: uppercase;
  box-shadow: var(--shadow-distance) var(--shadow-distance) 0px var(--border-color);
  transition: all 0.1s ease;
  font-family: var(--font-mono);
}

button:hover {
  transform: translate(-1px, -1px);
  box-shadow: calc(var(--shadow-distance) + 1px) calc(var(--shadow-distance) + 1px) 0px var(--border-color);
}

button:active {
  transform: translate(2px, 2px);
  box-shadow: 1px 1px 0px var(--border-color);
}

button.primary {
  background-color: var(--primary-color);
  color: white;
}

button.secondary {
  background-color: var(--secondary-color);
  color: black;
}

button.danger {
  background-color: var(--danger-color);
  color: white;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  background-color: #ddd;
  box-shadow: none;
  transform: none;
}

button.small-btn {
    padding: 0.3rem 0.6rem;
    font-size: 0.8rem;
}

.brutal-btn-lg {
  padding: 0.8rem 2rem;
  font-size: 1.1rem;
}

/* Inputs */
input[type="text"],
input[type="password"],
input[type="number"],
input[type="date"],
select,
textarea {
  width: 100%;
  padding: 0.8rem;
  background-color: white;
  border: var(--border-width) solid var(--border-color);
  color: var(--text-color);
  border-radius: 0;
  font-family: var(--font-mono);
  font-size: 1rem;
  box-sizing: border-box;
  box-shadow: 2px 2px 0px var(--border-color);
  transition: focus 0.2s;
}

input:focus, select:focus, textarea:focus {
  outline: none;
  background-color: #e0f2fe;
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

.title-badge {
    background: black;
    color: white;
    padding: 4px 8px;
    font-weight: bold;
    font-family: var(--font-mono);
    margin-bottom: 10px;
    transform: rotate(-2deg);
}

.home-title {
  font-size: 4rem;
  font-weight: 900;
  margin-bottom: 3rem;
  letter-spacing: 2px;
}

.home-actions {
  display: flex;
  gap: 2rem;
}

.action-card {
  position: relative;
  min-width: 12rem;
  background-color: white;
  padding: 2rem;
  border: var(--border-width) solid var(--border-color);
  cursor: pointer;
  transition: transform 0.1s;
  box-shadow: 8px 8px 0px var(--border-color);
}

.action-card.color-cyan { background-color: var(--cyan-color); }
.action-card.color-yellow { background-color: var(--secondary-color); }

.action-card:hover {
  transform: translate(-4px, -4px);
  box-shadow: 12px 12px 0px var(--border-color);
}

.action-card:active {
  transform: translate(4px, 4px);
  box-shadow: 0px 0px 0px var(--border-color);
}

.action-card .icon {
    margin-bottom: 1rem;
}

.action-card h2 {
  font-size: 1.5rem;
  margin: 0.5rem 0;
  text-transform: uppercase;
  font-weight: 800;
}

.action-card p {
    font-family: var(--font-mono);
    font-size: 0.9rem;
}

/* View Header */
.view-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding-bottom: 1rem;
  margin-bottom: 1.5rem;
  border-bottom: var(--border-width) solid var(--border-color);
}

.back-button {
    border: var(--border-width) solid var(--border-color);
    background: transparent;
    box-shadow: none;
}
.back-button:hover {
    background: black;
    color: white;
}

/* Stepper */
.stepper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 90%;
  margin: 1rem auto 2rem;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
  z-index: 2;
}

.step-box {
  width: 40px;
  height: 40px;
  background-color: white;
  border: var(--border-width) solid var(--border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 900;
  font-size: 1.2rem;
  transition: all 0.3s;
}

.step.active .step-box {
  background-color: var(--primary-color);
  color: white;
  box-shadow: 4px 4px 0px var(--border-color);
}

.step.completed .step-box {
  background-color: var(--success-color);
  color: black;
}

.step-title {
  margin-top: 8px;
  font-weight: bold;
  font-size: 0.9rem;
  background: white;
  padding: 2px 4px;
  border: 1px solid transparent;
}

.step.active .step-title {
    border: 2px solid black;
    box-shadow: 2px 2px 0 black;
}

.step-line {
  flex-grow: 1;
  height: var(--border-width);
  background-color: #ccc;
  margin: 0 -10px;
  transform: translateY(-12px);
  z-index: 1;
}

.step-line.completed {
  background-color: var(--border-color);
}

/* File Table */
.file-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background: #eee;
  border-bottom: var(--border-width) solid var(--border-color);
}

.badge-count {
    background: black;
    color: white;
    padding: 2px 6px;
    font-weight: bold;
    font-family: var(--font-mono);
}

.table-container {
    max-height: 400px;
    overflow-y: auto;
}

.file-table {
  width: 100%;
  border-collapse: collapse;
  font-family: var(--font-mono);
}

.file-table th {
  text-align: left;
  padding: 0.8rem;
  background: #fff;
  border-bottom: var(--border-width) solid var(--border-color);
  text-transform: uppercase;
  font-weight: 800;
  border-right: 1px solid #000;
}

.file-table td {
  padding: 0.6rem 0.8rem;
  border-bottom: 1px solid #000;
  border-right: 1px solid #000;
}

.file-table tr:hover {
    background-color: #fdf6e3;
}

.row-selected {
    background-color: #e0e7ff !important;
}

.dir-link {
    font-weight: bold;
    color: var(--primary-color);
    text-decoration: underline;
}

/* Switch Toggle */
.switch-group {
    justify-content: space-between;
}
input[type="checkbox"].toggle {
  -webkit-appearance: none;
  width: 50px;
  height: 26px;
  background-color: white;
  border: var(--border-width) solid var(--border-color);
  border-radius: 0;
  position: relative;
  cursor: pointer;
}
input[type="checkbox"].toggle::after {
  content: '';
  position: absolute;
  width: 16px;
  height: 16px;
  background-color: black;
  top: 2px;
  left: 3px;
  transition: left 0.2s;
}
input[type="checkbox"].toggle:checked {
  background-color: var(--success-color);
}
input[type="checkbox"].toggle:checked::after {
  left: 25px;
}

/* Progress Bar */
.progress-container {
    border: var(--border-width) solid var(--border-color);
    padding: 2px;
    background: white;
    box-shadow: 4px 4px 0 black;
}

progress {
  width: 100%;
  height: 20px;
  -webkit-appearance: none;
  background-color: transparent;
  border: none;
}

progress::-webkit-progress-bar {
  background-color: transparent;
}

progress::-webkit-progress-value {
  background-color: var(--primary-color);
  /* striped background for retro feel */
  background-image: linear-gradient(
    45deg,
    rgba(255, 255, 255, 0.15) 25%,
    transparent 25%,
    transparent 50%,
    rgba(255, 255, 255, 0.15) 50%,
    rgba(255, 255, 255, 0.15) 75%,
    transparent 75%,
    transparent
  );
  background-size: 1rem 1rem;
}

.progress-stage {
    font-family: var(--font-mono);
    font-weight: bold;
    margin-top: 10px;
    text-transform: uppercase;
}

.log-box.neo-log {
    border: var(--border-width) solid var(--border-color);
    background: black;
    color: var(--success-color);
    font-family: "Courier New", monospace;
    padding: 10px;
    margin-top: 10px;
    height: 200px;
    overflow-y: auto;
}

/* Profile Section */
.profile-buttons {
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    gap: 10px;
    margin-bottom: 10px;
}
.profile-buttons button {

  width: 5rem;
  height: 3em;
    background: #f1f1f1;
    box-shadow: 2px 2px 0 #000;
    font-size: 0.9rem;
    margin-bottom: 5px;
}
.profile-buttons button:hover {
    background: var(--cyan-color);
}

/* Modals */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6); /* 加深的遮罩 */
  backdrop-filter: grayscale(100%) contrast(120%); /* 粗糙的滤镜效果 */
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.neo-modal {
    background: white;
    border: 4px solid black;
    box-shadow: 10px 10px 0 black;
    padding: 0;
    max-width: 450px;
    width: 90%;
    position: relative;
}

.modal-header-strip {
    background: black;
    color: white;
    font-weight: bold;
    padding: 5px 15px;
    font-family: var(--font-mono);
}
.modal-header-strip.success-strip { background: var(--success-color); color: black; }
.modal-header-strip.error-strip { background: var(--danger-color); color: white; }
.modal-header-strip.warning-strip { background: var(--secondary-color); color: black; }

.neo-modal h3 {
    margin: 20px;
    font-size: 1.5rem;
    text-transform: uppercase;
}
.neo-modal p {
    margin: 0 20px 20px 20px;
}
.modal-input {
    margin: 0 20px 20px 20px;
}
.modal-actions {
    padding: 20px;
    background: #eee;
    border-top: 3px solid black;
    display: flex;
    justify-content: flex-end;
    gap: 10px;
}

/* History Item */
.history-item {
    padding: 1rem;
    border-bottom: 1px solid #000;
    transition: background 0.2s;
}
.history-item:hover {
    background-color: var(--cyan-color);
}
.history-item-wrapper {
    border: 2px solid black;
    margin-bottom: -2px; /* Collapse borders */
}
.backup-details {
    background: #333;
    color: white;
    padding: 1rem;
    font-family: var(--font-mono);
}
.backup-details button {
    width: 100%;
    margin-top: 10px;
}

/* Filter Grid */
.filter-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
}

.filter-group label {
    display: block;
    font-weight: bold;
    margin-bottom: 5px;
    text-transform: uppercase;
    font-size: 0.8rem;
}

.size-input {
    display: flex;
    gap: 5px;
}


</style>