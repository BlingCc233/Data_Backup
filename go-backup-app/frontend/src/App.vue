<template>
  <div id="app-container">
    <!-- Main View: Home Screen -->
    <main class="main-content" v-if="currentScreen === 'home'">
      <div class="home-screen">
        <h1 class="home-title">CcBackup</h1>
        <div class="home-actions">
          <div class="action-card" @click="navigateTo('backup')">
            <div class="icon">
              <svg height="4rem" width="4rem" version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg"
                   xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 511.998 511.998" xml:space="preserve"
                   fill="#000000"><g id="SVGRepo_bgCarrier" stroke-width="0"></g>
                <g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g>
                <g id="SVGRepo_iconCarrier"> <g> <polygon style="fill:#CFF09E;"
                                                          points="351.38,16.284 347.586,16.284 347.586,122.034 464.78,122.034 "></polygon>
                  <polygon style="fill:#CFF09E;"
                           points="44.666,208.782 44.666,306.488 184.933,306.488 184.933,370.87 326.573,255.999 184.933,141.13 184.933,208.782 "></polygon> </g>
                  <g> <path style="fill:#507C5C;"
                            d="M467.331,202.728c-8.994,0-16.284,7.29-16.284,16.284v260.42H105.826v-156.66h62.825v48.098 c0,6.279,3.61,12,9.279,14.7c2.231,1.063,4.623,1.584,7.002,1.584c3.667,0,7.302-1.238,10.259-3.636l141.64-114.871 c3.812-3.091,6.027-7.738,6.027-12.648c0-4.91-2.215-9.555-6.027-12.648l-141.64-114.871c-4.877-3.955-11.593-4.753-17.263-2.052 c-5.668,2.702-9.279,8.42-9.279,14.7v51.368H44.668c-8.994,0-16.284,7.29-16.284,16.284v97.706c0,8.994,7.29,16.284,16.284,16.284 h28.59v172.944c0,8.994,7.29,16.284,16.284,16.284h377.789c8.994,0,16.284-7.29,16.284-16.284V219.012 C483.615,210.018,476.324,202.728,467.331,202.728z M60.952,225.066h123.983c8.994,0,16.284-7.29,16.284-16.284v-33.48 l99.503,80.697l-99.503,80.697v-30.208c0-8.994-7.29-16.284-16.284-16.284H60.952V225.066z"></path>
                    <path style="fill:#507C5C;"
                          d="M89.542,138.318c8.994,0,16.284-7.29,16.284-16.284V32.568h225.476v89.466 c0,8.994,7.29,16.284,16.284,16.284H464.78c6.689,0,12.698-4.092,15.151-10.316c2.452-6.224,0.848-13.315-4.045-17.878 L362.488,4.376c-0.124-0.116-0.259-0.217-0.386-0.327c-0.192-0.168-0.383-0.335-0.583-0.493c-0.2-0.16-0.405-0.309-0.612-0.458 c-0.217-0.156-0.432-0.311-0.656-0.458c-0.199-0.129-0.402-0.248-0.606-0.368c-0.246-0.147-0.493-0.29-0.747-0.423 c-0.191-0.099-0.386-0.187-0.58-0.28c-0.275-0.13-0.55-0.261-0.835-0.375c-0.189-0.077-0.383-0.14-0.575-0.21 c-0.295-0.106-0.588-0.213-0.889-0.303c-0.207-0.062-0.417-0.107-0.625-0.161c-0.291-0.073-0.58-0.153-0.878-0.21 c-0.267-0.052-0.537-0.085-0.806-0.124c-0.246-0.034-0.489-0.081-0.738-0.104C352.445,0.029,351.914,0,351.38,0h-3.794H89.542 c-8.994,0-16.284,7.29-16.284,16.284v105.75C73.258,131.028,80.548,138.318,89.542,138.318z M363.87,105.75V50.197l59.572,55.553 H363.87z"></path> </g> </g></svg>
            </div>
            <h2>备份</h2>
            <p>创建新的数据备份</p>
          </div>
          <div class="action-card" @click="navigateTo('restore')">
            <div class="icon">
              <svg height="4rem" width="4rem" version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg"
                   xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 512.002 512.002" xml:space="preserve"
                   fill="#000000"><g id="SVGRepo_bgCarrier" stroke-width="0"></g>
                <g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g>
                <g id="SVGRepo_iconCarrier"> <g> <polygon style="fill:#9ed7f0;"
                                                          points="351.384,16.284 347.589,16.284 347.589,122.034 464.784,122.034 "></polygon>
                  <polygon style="fill:#9ed7f0;"
                           points="326.576,303.218 326.576,205.512 186.309,205.512 186.309,141.13 44.669,255.999 186.309,370.87 186.309,303.218 "></polygon> </g>
                  <g> <path style="fill:#50597c;"
                            d="M467.334,202.728c-8.994,0-16.284,7.29-16.284,16.284v260.42H105.829V326.566l70.225,56.953 c2.957,2.399,6.592,3.636,10.259,3.636c2.379,0,4.773-0.521,7.002-1.584c5.668-2.702,9.279-8.42,9.279-14.7v-51.368h123.983 c8.994,0,16.284-7.29,16.284-16.284v-97.706c0-8.994-7.29-16.284-16.284-16.284H202.593v-48.1c0-6.279-3.61-12-9.279-14.7 c-5.665-2.702-12.384-1.902-17.263,2.052L34.412,243.353c-3.812,3.091-6.027,7.738-6.027,12.648c0,4.91,2.215,9.555,6.027,12.648 l38.847,31.505v195.564c0,8.994,7.29,16.284,16.284,16.284h377.789c8.994,0,16.284-7.29,16.284-16.284V219.012 C483.618,210.018,476.327,202.728,467.334,202.728z M170.027,175.302v30.208c0,8.994,7.29,16.284,16.284,16.284h123.983v65.138 H186.311c-8.994,0-16.284,7.29-16.284,16.284v33.48l-99.503-80.697L170.027,175.302z"></path>
                    <path style="fill:#50597c;"
                          d="M89.545,138.318c8.994,0,16.284-7.29,16.284-16.284V32.568h225.476v89.466 c0,8.994,7.29,16.284,16.284,16.284h117.194c6.689,0,12.698-4.092,15.151-10.316c2.452-6.224,0.848-13.315-4.045-17.878 L362.491,4.376c-0.124-0.116-0.259-0.217-0.386-0.327c-0.192-0.168-0.383-0.335-0.583-0.493c-0.2-0.16-0.405-0.309-0.612-0.458 c-0.217-0.156-0.432-0.311-0.656-0.458c-0.199-0.129-0.402-0.248-0.606-0.368c-0.246-0.147-0.493-0.29-0.747-0.423 c-0.191-0.099-0.386-0.187-0.58-0.28c-0.275-0.13-0.55-0.261-0.835-0.375c-0.189-0.077-0.383-0.14-0.575-0.21 c-0.295-0.106-0.588-0.213-0.889-0.303c-0.207-0.062-0.417-0.107-0.625-0.161c-0.292-0.073-0.58-0.153-0.878-0.21 c-0.267-0.052-0.537-0.085-0.806-0.124c-0.246-0.034-0.489-0.081-0.738-0.104C352.448,0.029,351.918,0,351.384,0h-3.794H89.545 c-8.994,0-16.284,7.29-16.284,16.284v105.75C73.261,131.028,80.551,138.318,89.545,138.318z M363.873,105.75V50.197l59.572,55.553 H363.873z"></path> </g> </g></svg>
            </div>
            <h2>恢复</h2>
            <p>从备份文件中恢复数据</p>
          </div>
        </div>
      </div>
    </main>

    <!-- Nested Views: Backup/Restore pages -->
    <main class="main-content" v-else>
      <div class="view-header">
        <button class="back-button" @click="navigateBack">← 返回首页</button>
        <h2 class="view-title">{{ viewTitle }}</h2>
      </div>

      <!-- Backup View -->
      <div v-if="currentScreen === 'backup'" class="view">
        <div class="stepper">
          <div class="step" :class="{ active: backupStep === 1, completed: backupStep > 1 }">
            <div class="step-circle">1</div>
            <div class="step-title">选择项目</div>
          </div>
          <div class="step-line" :class="{ completed: backupStep > 1 }"></div>
          <div class="step" :class="{ active: backupStep === 2, completed: backupStep > 2 }">
            <div class="step-circle">2</div>
            <div class="step-title">高级筛选</div>
          </div>
          <div class="step-line" :class="{ completed: backupStep > 2 }"></div>
          <div class="step" :class="{ active: backupStep === 3 }">
            <div class="step-circle">3</div>
            <div class="step-title">加密与执行</div>
          </div>
        </div>

        <!-- Step 1: Select Paths & Files -->
        <div v-if="backupStep === 1">
          <div class="card">
            <h3>Step 1: 选择要备份的文件或文件夹</h3>

            <!-- Profile Section -->
            <div class="profile-section">
              <h4 class="profile-title">通过档案快速选择</h4>
              <p class="description">选择一个预设或自定义的档案来快速添加常用文件夹。</p>
              <div class="profile-actions">
                <div class="profile-buttons">
                  <button v-for="profile in profiles" :key="profile.id" @click="applyProfile(profile)"
                          class="profile-btn">
                    {{ profile.name }}
                  </button>
                </div>
                <button @click="openSaveProfileModal" :disabled="selectedBackupFileCount === 0">
                  将当前选择保存为档案
                </button>
              </div>
            </div>
            <hr class="card-divider">

            <p class="description">或者，您可以手动多选文件或文件夹。选择后，它们将显示在下面的列表中。</p>
            <div class="action-bar-left">
              <button @click="selectBackupSources('files')">选择文件</button>
              <button @click="selectBackupSources('dirs')">选择文件夹</button>
            </div>
          </div>

          <div class="card file-list-card" v-if="backupFiles.length > 0">
            <div class="file-list-header">
              <!-- Breadcrumbs -->
              <div class="breadcrumbs">
                <span v-for="(part, index) in pathStack" :key="part.path" @click="navigateToBreadcrumb(index)"
                      class="breadcrumb-item">
                  {{ part.name }}
                </span>
              </div>
              <span>共 {{ selectedBackupFileCount }} 项已选</span>
            </div>

            <!-- Loading indicator -->
            <div v-if="isBrowsing" class="loading-overlay">
              <p>正在加载...</p>
            </div>

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
                <th>权限</th>
              </tr>
              </thead>
              <tbody>
              <!-- Iterating over sortedCurrentViewItems now -->
              <tr v-for="file in sortedCurrentViewItems" :key="file.path">
                <td><input type="checkbox" v-model="file.selected" @change="updateSelection(file)"></td>
                <td class="col-name">
                  <span class="file-icon">{{ file.isDir ? '📁' : '📄' }}</span>
                  <!-- Click handler on the name for navigation -->
                  <span :class="{ 'dir-link': file.isDir }" @click="file.isDir ? enterDirectory(file) : null">
                     {{ file.name }}
                  </span>
                </td>
                <td>{{ formatSize(file.size) }}</td>
                <td>{{ formatDate(file.modTime) }}</td>
                <td>{{ file.mode }}</td>
              </tr>
              <tr v-if="!isBrowsing && sortedCurrentViewItems.length === 0">
                <td colspan="5" class="empty-dir-msg">这个文件夹是空的。</td>
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
            <strong style="color: #f6ad55;">如果不需要筛选，请直接点击下一步。</strong>
            <div class="filter-grid">
              <div class="filter-group">
                <label>包含名称 (e.g. `*.log`, `data*`)</label>
                <textarea v-model="filters.includeNames" placeholder="一行一个匹配模式"></textarea>
              </div>
              <div class="filter-group">
                <label>排除名称 (e.g. `*.tmp`, `cache*`)</label>
                <textarea v-model="filters.excludeNames" placeholder="一行一个匹配模式"></textarea>
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
                  <input type="password" v-model="encryption.password" placeholder="输入一个强密码"/>
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
                <input v-model="backupDest" readonly type="text" placeholder="选择一个目录来保存备份文件"/>
                <button @click="selectDestDir">选择</button>
              </div>
            </div>
            <div class="action-bar">
              <button @click="backupStep = 2">上一步</button>
              <button class="primary" @click="doBackup"
                      :disabled="!backupDest || (encryption.enabled && !encryption.password)">
                开始备份
              </button>
            </div>
          </div>
          <!-- Backup Progress View -->
          <div v-if="currentScreen === 'backup' && backupStep === 3 && inProgress" class="progress-view">
            <h3>正在备份...</h3>
            <progress :value="progress.value" :max="100"></progress>
            <div class="progress-stage">{{ getProgressStageText() }}</div>
            <p class="status">{{ statusMessage }}</p>
            <div class="action-bar">
              <button class="danger" @click="stopOperation">停止备份</button>
            </div>
          </div>
        </div>
      </div>

      <!-- Restore View -->
      <div v-if="currentScreen === 'restore'" class="view">
        <div v-if="!inProgress">
          <div class="card">
            <h3>从最近的备份恢复</h3>
            <p class="description">点击一项来选择它作为恢复源。</p>

            <div class="backup-history-list">
              <div v-for="item in backupHistory" :key="item.ID"
                   class="history-item-wrapper"
                   :class="{ 'expanded': expandedBackupId === item.ID }">
                <div class="history-item" @click="toggleBackupDetails(item)">
                  <div class="history-item-main">
                    <span class="history-file">{{ item.FileName }}</span>
                    <span class="history-date">{{ formatDate(item.CreatedAt) }}</span>
                  </div>
                  <small class="history-path">{{ item.BackupPath }}</small>
                </div>
                <!-- Expanded Details View -->
                <div v-if="expandedBackupId === item.ID" class="backup-details">
                  <h4>包含的文件/文件夹:</h4>
                  <ul>
                    <li v-for="(path, index) in item.SourcePaths.split('\n').filter(p => p)" :key="index">
                      {{ path }}
                    </li>
                  </ul>
                </div>
              </div>
              <div v-if="!backupHistory.length" class="history-empty">
                暂无备份记录。
              </div>
            </div>
          </div>
          <div class="card">
            <h3>手动选择</h3>
            <div class="input-group">
              <label>备份文件:</label>
              <input v-model="restoreFile" readonly type="text" placeholder="选择一个 .qbak 文件"/>
              <button @click="selectRestoreFileManually">选择</button>
            </div>
            <div class="input-group">
              <label>恢复到目录:</label>
              <input v-model="restoreDir" readonly type="text" placeholder="选择恢复文件的位置"/>
              <button @click="selectRestoreDir">选择</button>
            </div>
          </div>
          <div class="action-bar">
            <button class="primary" @click="doRestore()" :disabled="!restoreFile || !restoreDir || inProgress">
              开始恢复
            </button>
          </div>
        </div>
        <!-- Restore Progress View -->
        <div v-if="currentScreen === 'restore' && inProgress" class="progress-view">
          <h3>正在恢复...</h3>
          <progress :value="progress.value" :max="100"></progress>
          <div class="progress-stage">{{ getProgressStageText() }}</div>
          <p class="status">{{ statusMessage }}</p>
          <div class="log-box" style="height: 300px; margin-bottom: 1rem;">
            <p v-for="(msg, index) in logMessages" :key="index">{{ msg }}</p>
          </div>
          <div class="action-bar">
            <button class="danger" @click="stopOperation">停止恢复</button>
          </div>
        </div>
      </div>
    </main>

    <!-- Global Log Section (only shown for backup) -->
    <!--    <div class="log-card-container" v-if="currentScreen === 'backup'">-->
    <div class="log-card-container">
      <div class="log-card">
        <h3>日志输出</h3>
        <p class="status">{{ statusMessage }}</p>
        <div class="log-box">
          <p v-for="(msg, index) in logMessages" :key="index">{{ msg }}</p>
        </div>
      </div>
    </div>

    <!-- Password Modal -->
    <div v-if="isPasswordModalVisible" class="modal-overlay">
      <div class="modal-content">
        <h3>需要密码</h3>
        <p>此备份文件已加密。请输入密码以继续恢复。</p>
        <div class="input-group modal-input">
          <label>密码:</label>
          <input
              type="password"
              v-model="restorePasswordInput"
              @keyup.enter="submitPasswordAndRetryRestore"
              placeholder="输入备份密码"
              ref="passwordInputRef"
          />
        </div>
        <div class="modal-actions">
          <button @click="cancelPasswordPrompt">取消</button>
          <button class="primary" @click="submitPasswordAndRetryRestore">提交</button>
        </div>
      </div>
    </div>

    <!-- Success Modal -->
    <div v-if="showSuccessModal" class="modal-overlay">
      <div class="modal-content success-modal">
        <h3>操作成功</h3>
        <p>{{ successMessage }}</p>
        <div class="modal-actions">
          <button class="primary" @click="closeSuccessModal">确定</button>
        </div>
      </div>
    </div>

    <!-- Error Modal -->
    <div v-if="showErrorModal" class="modal-overlay">
      <div class="modal-content error-modal">
        <h3>操作失败</h3>
        <p>错误信息：{{ successMessage }}</p>
        <div class="modal-actions">
          <button class="primary" @click="closeErrModal">确定</button>
        </div>
      </div>
    </div>

    <!-- Conflict Modal -->
    <div v-if="isConflictModalVisible" class="modal-overlay">
      <div class="modal-content">
        <h3>文件冲突</h3>
        <p>目标位置已存在文件：</p>
        <p><strong>{{ conflictInfo.path }}</strong></p>
        <p>您希望如何处理？</p>
        <div class="modal-actions">
          <button @click="resolveConflict('skip')">跳过</button>
          <button @click="resolveConflict('keep_both')">保留两者</button>
          <button class="primary" @click="resolveConflict('overwrite')">覆盖</button>
        </div>
      </div>
    </div>

    <!-- Profile Save Modal -->
    <div v-if="isProfileModalVisible" class="modal-overlay">
      <div class="modal-content">
        <h3>保存为新档案</h3>
        <p>为当前选择的 {{ selectedBackupFileCount }} 个项目输入一个档案名称。</p>
        <div class="input-group modal-input">
          <label>档案名称:</label>
          <input
              type="text"
              v-model="newProfileName"
              @keyup.enter="confirmSaveProfile"
              placeholder="例如: '我的工作文档'"
              ref="profileNameInputRef"
          />
        </div>
        <div class="modal-actions">
          <button @click="closeSaveProfileModal">取消</button>
          <button class="primary" @click="confirmSaveProfile" :disabled="!newProfileName.trim()">保存</button>
        </div>
      </div>
    </div>

  </div>

</template>

<script setup>
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
    statusMessage.value = data; // Also update status for real-time feedback
    if (logMessages.value.length > 200) logMessages.value.pop();
  });
  EventsOn("progress_update", (p) => {
    statusMessage.value = p.message;
  });
  EventsOn("conflict_detected", (data) => {
    // data 包含 { path: "...", requestID: "..." }
    conflictInfo.path = data.path;
    conflictInfo.requestID = data.requestID;
    isConflictModalVisible.value = true;
  });

  fetchBackupHistory();
  fetchProfiles();
});


// --- Global State ---
const currentScreen = ref('home'); // 'home', 'backup', 'restore'
const inProgress = ref(false);
const statusMessage = ref('准备就绪。');
const logMessages = ref([]);
const progress = reactive({value: 0, max: 100});
const isPasswordModalVisible = ref(false);
const restorePasswordInput = ref('');
const passwordInputRef = ref(null);


// 添加进度阶段状态
const progressStage = ref(''); // 'compressing', 'encrypting', 'archiving', 'decrypting', 'decompressing', 'restoring'
const progressStartTime = ref(0);

// --- Navigation ---
const viewTitle = computed(() => {
  if (currentScreen.value === 'backup') return '创建新备份';
  if (currentScreen.value === 'restore') return '从备份恢复';
  return 'GoBackup';
});

function navigateTo(screen) {
  currentScreen.value = screen;
  // Reset states when entering a new screen
  if (screen === 'backup') resetBackupState();
  if (screen === 'restore') resetRestoreState();
}

// Back button always returns to home
function navigateBack() {
  currentScreen.value = 'home';
  if (inProgress.value) {
    stopOperation();
  }
}


// --- Backup State & Logic ---
const backupStep = ref(1);
const backupFiles = ref([]);
const backupDest = ref('');
const sort = reactive({key: 'name', order: 'asc'});
const filters = ref({
  includeNames: '', excludeNames: '', minSizeValue: 0, minSizeUnit: 'Bytes',
  maxSizeValue: 0, maxSizeUnit: 'Bytes', newerThan: null, olderThan: null,
});
const encryption = reactive({enabled: false, password: '', algorithm: 'AES-256'});

const pathStack = ref([{name: '根目录', path: 'root'}]); // Breadcrumb stack
const currentViewItems = ref([]); // Items currently in the table
const isBrowsing = ref(false); // Loading indicator state
const fileSelectionMap = reactive(new Map()); // Master map for selection state { path: boolean }

// --- Profile State & Logic ---
const profiles = ref([]);
const isProfileModalVisible = ref(false);
const newProfileName = ref('');
const profileNameInputRef = ref(null);

async function fetchProfiles() {
  try {
    profiles.value = await GetProfiles();
  } catch (e) {
    statusMessage.value = `无法加载档案: ${e}`;
  }
}

async function applyProfile(profile) {
  // Clear existing selections first
  backupFiles.value = [];
  fileSelectionMap.clear();
  pathStack.value = [{name: '根目录', path: 'root'}];
  currentViewItems.value = [];

  const paths = profile.paths.split('\n').filter(p => p.trim() !== '');
  if (paths.length > 0) {
    try {
      statusMessage.value = `正在从档案 '${profile.name}' 加载...`;
      const metadata = await GetFileMetadata(paths);
      metadata.forEach(m => {
        if (!backupFiles.value.some(f => f.path === m.path)) {
          backupFiles.value.push({...m, selected: true});
        }
        fileSelectionMap.set(m.path, true);
      });
      await loadDirectoryView('root');
    } catch (error) {
      statusMessage.value = `从档案加载时出错: ${error}`;
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
    statusMessage.value = "档案名称不能为空。";
    return;
  }
  const selectedPaths = Array.from(fileSelectionMap.entries())
      .filter(([, selected]) => selected)
      .map(([path]) => path);

  if (selectedPaths.length === 0) {
    statusMessage.value = "没有选择任何项目来保存。";
    return;
  }

  try {
    await CreateProfile(name, selectedPaths);
    statusMessage.value = `档案 '${name}' 已保存。`;
    await fetchProfiles(); // Refresh the list of profiles
    closeSaveProfileModal();
  } catch (error) {
    statusMessage.value = `保存档案失败: ${error}`;
    successMessage.value = `保存档案失败: ${error}`;
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

  pathStack.value = [{name: '根目录', path: 'root'}];
  currentViewItems.value = [];
  fileSelectionMap.clear();

}

// 添加获取进度阶段文本的函数
function getProgressStageText() {
  switch (progressStage.value) {
    case 'compressing':
      return '压缩中...';
    case 'encrypting':
      return '加密中...';
    case 'archiving':
      return '存档中...';
    case 'decrypting':
      return '解密中...';
    case 'decompressing':
      return '解压缩中...';
    case 'restoring':
      return '恢复文件中...';
    default:
      return '处理中...';
  }
}

async function selectBackupSources(type) {
  try {
    const paths = await SelectFiles(type === 'dirs');
    if (paths && paths.length > 0) {
      const metadata = await GetFileMetadata(paths);
      metadata.forEach(m => {
        // Add to top-level list if not already there
        if (!backupFiles.value.some(f => f.path === m.path)) {
          backupFiles.value.push({...m, selected: true});
        }
        // Add to selection map
        fileSelectionMap.set(m.path, true);
      });
      // Refresh the root view
      await loadDirectoryView('root');
    }
  } catch (error) {
    statusMessage.value = `Error selecting sources: ${error}`;
  }
}

async function loadDirectoryView(path) {
  isBrowsing.value = true;
  statusMessage.value = `正在加载 ${path}...`;
  try {
    let items;
    if (path === 'root') {
      // Root view shows the initial selected files/folders
      items = [...backupFiles.value];
    } else {
      // Fetch contents from backend for subdirectories
      items = await ListDirectory(path);
    }

    // Sync selection state from the master map
    currentViewItems.value = items.map(item => ({
      ...item,
      selected: fileSelectionMap.get(item.path) || false
    }));

  } catch (error) {
    statusMessage.value = `加载目录失败: ${error}`;
    currentViewItems.value = []; // Clear view on error
  } finally {
    isBrowsing.value = false;
    statusMessage.value = '准备就绪。';
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
    // Put directories first
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

// 添加计算对数进度的函数
function calculateLogProgress(elapsedTime) {
  const maxTime = 200 * 1000; // 毫秒
  // 使用对数函数实现由快到慢的进度
  if (elapsedTime >= maxTime) return 90;

  // 对数进度计算 (以10为底)
  const progress = Math.log10(1 + 9 * (elapsedTime / maxTime)) * 100;
  return Math.min(99, Math.max(0, progress));
}

// 添加更新进度的函数
function updateProgress() {
  if (!inProgress.value || progressStartTime.value === 0) return;


  if (isConflictModalVisible.value) {
    progress.value = progress.value;
  } else {
    const elapsed = Date.now() - progressStartTime.value;
    progress.value = calculateLogProgress(elapsed);
  }

  // 特殊处理加密情况下的进度提示切换
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

  // 根据阶段更新状态消息
  if (progressStage.value === 'compressing') {
    statusMessage.value = "正在压缩...";
  } else if (progressStage.value === 'encrypting') {
    statusMessage.value = "正在加密...";
  } else if (progressStage.value === 'archiving') {
    statusMessage.value = "正在存档...";
  } else if (progressStage.value === 'decrypting') {
    statusMessage.value = "正在解密...";
  } else if (progressStage.value === 'decompressing') {
    statusMessage.value = "正在解压缩...";
  } else if (progressStage.value === 'restoring') {
    statusMessage.value = "正在恢复文件...";
  }

  // 继续更新进度直到操作完成
  if (inProgress.value) {
    setTimeout(updateProgress, 100);
  }
}

async function doBackup() {
  const selectedPaths = Array.from(fileSelectionMap.entries())
      .filter(([, selected]) => selected)
      .map(([path]) => path);
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
      UseCompression: true,
      useEncryption: encryption.enabled,
      encryptionAlgorithm: encryption.algorithm,
      encryptionPassword: encryption.password,
    });

    statusMessage.value = result;

    // 备份完成，进度条拉满
    progress.value = 100;

    // 显示成功弹窗
    if (result.includes("成功")) {
      successMessage.value = "备份完成！";
      showSuccessModal.value = true;
    }
  } catch (error) {
    successMessage.value = `${error}`;
    showErrorModal.value = true;
    statusMessage.value = `错误: ${error}`;
  } finally {
    inProgress.value = false;
    await fetchBackupHistory();
  }
}

// --- Restore State & Logic ---
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
  expandedBackupId.value = null; // Reset expanded item
  fetchBackupHistory();
}

function toggleBackupDetails(item) {
  // Always set the selected file for restoration
  restoreFile.value = item.BackupPath;

  // Toggle the expanded view
  if (expandedBackupId.value === item.ID) {
    expandedBackupId.value = null; // Collapse if it's already open
  } else {
    expandedBackupId.value = item.ID; // Expand the clicked item
  }
}

async function fetchBackupHistory() {
  try {
    backupHistory.value = await GetBackupHistory();
  } catch (e) {
    statusMessage.value = `无法加载备份历史: ${e}`;
  }
}

async function selectRestoreFileManually() {
  const path = await SelectFiles(false); // false for file, returns array
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
    statusMessage.value = "请选择备份文件和恢复目录。";
    return;
  }

  inProgress.value = true;
  statusMessage.value = "恢复进行中...";
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

    // 显示成功弹窗
    if (result.includes("成功")) {
      successMessage.value = "恢复完成！";
      showSuccessModal.value = true;
      await OpenInExplorer(restoreDir.value);
    }
  } catch (error) {
    if (typeof error === 'string' && error.includes("password_required")) {
      inProgress.value = false;
      statusMessage.value = "此备份已加密，请输入密码。";
      isPasswordModalVisible.value = true;
      await nextTick();
      passwordInputRef.value?.focus();
    } else {
      statusMessage.value = `错误: ${error}`;
      showErrorModal.value = true;
      if (cleanPassword !== '' && error.includes("EOF")) {
        successMessage.value = `密码错误请重试`;
      } else {
        successMessage.value = `${error}`;
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
    successMessage.value = `密码不能为空`;
    showErrorModal.value = true;
    return;
  }
  isPasswordModalVisible.value = false;
  inProgress.value = true;
  doRestore(restorePasswordInput.value); // Retry with password
  restorePasswordInput.value = '';
}

function cancelPasswordPrompt() {
  isPasswordModalVisible.value = false;
  statusMessage.value = "未提供密码，恢复已取消。";
  inProgress.value = false;
  restorePasswordInput.value = '';
}

// --- Common & Helpers ---
function stopOperation() {
  StopOperation();
  inProgress.value = false;
  statusMessage.value = "操作已停止。";
  progressStartTime.value = 0;
  progress.value = 0; // 重置进度条
  progressStage.value = ''; // 清除阶段状态

  if (currentScreen.value === 'backup') {
    backupStep.value = Math.max(1, backupStep.value - 1);
  }
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

// 添加成功弹窗状态
const showSuccessModal = ref(false);
const successMessage = ref('');
const showErrorModal = ref(false);

// 添加关闭成功弹窗并返回首页的函数
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

// --- 冲突模态框状态 ---
const isConflictModalVisible = ref(false);
const conflictInfo = reactive({
  path: '',
  requestID: '',
});


// --- 冲突解决逻辑 ---
async function resolveConflict(resolution) {
  if (!conflictInfo.requestID) return;

  try {
    await ResolveConflict(conflictInfo.requestID, resolution);
    // 成功发送后，关闭模态框并重置状态
    isConflictModalVisible.value = false;
    conflictInfo.path = '';
    conflictInfo.requestID = '';
  } catch (error) {
    // 处理错误，例如显示一个错误消息
    statusMessage.value = `解决冲突失败: ${error}`;
    showErrorModal.value = true;
    successMessage.value = `解决冲突失败: ${error}`;
    // 也许也应该关闭模态框
    isConflictModalVisible.value = false;
  }
}

</script>

<style>
/* Keeping original styles and adding/modifying new ones */
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
  flex-direction: column;
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
  justify-content: space-evenly;
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
  gap: 3rem;
}

.action-card {
  min-width: 10rem;
  background-color: var(--card-bg);
  padding: 2rem 3rem;
  border-radius: 12px;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  border: 1px solid var(--border-color);
}

.action-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.2);
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
.file-list-card {
  padding: 0;
}

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
  min-height: 2rem;
  max-height: 3rem;
}

.file-table tbody tr:last-child td {
  border-bottom: none;
}

.file-table th {
  color: var(--text-color-light);
  font-size: 0.9rem;
  text-transform: uppercase;
}

.file-table th.sortable {
  cursor: pointer;
}

.file-table th.sortable:hover {
  color: var(--text-color);
}

.file-table .col-checkbox {
  width: 40px;
  text-align: center;
}

.file-table .col-name {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-height: 3rem;
  max-width: 15rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
}

.file-icon {
  font-size: 1.2rem;
}

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

.history-date {
  color: var(--text-color-light);
  font-size: 0.9rem;
}

.history-path {
  color: var(--text-color-light);
  font-size: 0.8rem;
}

.history-empty {
  text-align: center;
  padding: 2rem;
  color: var(--text-color-light);
}

/* Progress View */
.progress-view {
  text-align: center;
  padding: 2rem;
}

.progress-view h3 {
  margin-top: 0;
}

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

progress::-webkit-progress-bar {
  background-color: var(--bg-color);
}

progress::-webkit-progress-value {
  background-color: var(--primary-color);
  transition: width 0.1s linear;
}

.progress-stage {
  margin-top: 0.5rem;
  font-size: 0.9rem;
  color: var(--text-color-light);
  font-style: italic;
}

/* Actions */
.action-bar {
  text-align: right;
  margin-top: 1.5rem;
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
}

.action-bar-left {
  text-align: left;
  display: flex;
  gap: 1rem;
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

button.danger {
  background-color: var(--danger-color);
  border-color: var(--danger-color);
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Log Section (at the bottom) */
.log-card-container {
  display: none;
  /* TODO */
  flex-shrink: 0;
  padding: 0 2rem 2rem 2rem;
}

.log-card {
  margin-top: 0;
  background-color: var(--card-bg);
  border-radius: 8px;
  padding: 1rem 1.5rem;
}

.log-card h3 {
  margin: 0;
  font-size: 1rem;
  color: var(--text-color-light);
}

.status {
  color: var(--text-color-light);
  font-style: italic;
  font-size: 0.9rem;
}

.log-box {
  height: 150px;
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

/* Modal Styles */
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
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.3);
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


/* Other component styles (from original) */
.filter-grid {
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
  display: grid;
  margin-top: 1.5rem;
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

.encryption-options {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--border-color);
}

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

/* Success Modal Styles */
.success-modal {
  text-align: center;
}

.success-modal h3 {
  color: #48bb78; /* Green color for success */
  font-size: 1.8rem;
  margin-bottom: 1rem;
}

.success-modal p {
  font-size: 1.2rem;
  margin-bottom: 2rem;
}

/* Error Modal Styles */
.error-modal {
  text-align: center;
}

.error-modal h3 {
  color: #bb4870; /* Green color for success */
  font-size: 1.8rem;
  margin-bottom: 1rem;
}

.error-modal p {
  font-size: 1.2rem;
  margin-bottom: 2rem;
}

/* Stepper Styles */
.stepper {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 80%;
  margin: 1rem auto 2rem;
  padding: 0 1rem;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  color: #a0aec0; /* Inactive color */
  transition: color 0.3s;
}

.step-circle {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: #4a5568;
  border: 2px solid #a0aec0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  margin-bottom: 0.5rem;
  transition: background-color 0.3s, border-color 0.3s;
}

.step-title {
  font-size: 0.9rem;
  font-weight: 500;
}

.step-line {
  flex-grow: 1;
  height: 2px;
  background-color: #4a5568;
  margin: 0 1rem;
  transform: translateY(-16px); /* Align with the middle of the circles */
  transition: background-color 0.3s;
}

/* Active & Completed States */
.step.active .step-circle {
  background-color: #4299e1; /* Active blue */
  border-color: #4299e1;
}

.step.active {
  color: #e2e8f0; /* Active text color */
}

.step.completed .step-circle {
  background-color: #48bb78; /* Completed green */
  border-color: #48bb78;
}

.step.completed {
  color: #c6f6d5; /* Completed text color */
}

.step-line.completed {
  background-color: #48bb78;
}

/* Restore Item Expansion Styles */
.history-item-wrapper {
  border-bottom: 1px solid #4a5568;
  transition: background-color 0.2s;
}

.history-item-wrapper:last-child {
  border-bottom: none;
}

.history-item {
  padding: 1rem;
  cursor: pointer;
}

.history-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.history-item-wrapper.expanded .history-item {
  background-color: #3b485b; /* A slightly different background when expanded */
}

.backup-details {
  background-color: #2d3748; /* Darker background for details */
  padding: 1rem 1.5rem;
  border-top: 1px solid #4a5568;
}

.backup-details h4 {
  margin-top: 0;
  margin-bottom: 0.75rem;
  color: #a0aec0;
  font-size: 0.9rem;
}

.backup-details ul {
  list-style-type: none;
  padding-left: 0;
  margin: 0;
}

.backup-details li {
  padding: 0.25rem 0;
  font-size: 0.85rem;
  color: #cbd5e0;
  word-break: break-all;
}

.breadcrumbs {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  margin-bottom: -1rem; /* Adjust to align with header */
  padding-bottom: 1rem;
}

.breadcrumb-item {
  color: #a0aec0;
  cursor: pointer;
  transition: color 0.2s;
}

.breadcrumb-item:hover {
  color: #63b3ed;
}

.breadcrumb-item::after {
  content: '>';
  margin: 0 0.5rem;
  color: #718096;
}

.breadcrumb-item:last-child {
  color: #e2e8f0;
  cursor: default;
}

.breadcrumb-item:last-child::after {
  content: '';
}

.dir-link {
  color: #63b3ed;
  cursor: pointer;
  text-decoration: none;
}

.dir-link:hover {
  text-decoration: underline;
}

.file-list-card {
  position: relative; /* Needed for loading overlay */
}

.loading-overlay {
  position: absolute;
  top: 50px; /* Below header */
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(30, 41, 59, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  color: #e2e8f0;
  font-size: 1.2rem;
}

.file-table.is-loading tbody {
  opacity: 0.3;
}

.empty-dir-msg {
  text-align: center;
  padding: 2rem;
  color: #a0aec0;
}

/* Basic styles for the new profile section */
.profile-section {
  margin-top: 1.5rem;
  margin-bottom: 1.5rem;
}

.profile-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.profile-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.profile-buttons {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
}

.profile-buttons button {
  padding: 0 0.6rem 0.4rem 0.6rem;
  background-color: var(--sidebar-bg);
  border: none;
  border-bottom: 1px solid #9ed7f0 !important;
  color: var(--text-color);
  border-radius: 0px;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.2s;
}

.profile-buttons button:hover {
  border-bottom: 1px solid #7393d5 !important;
}

.card-divider {
  border: none;
  height: 1px;
  background-color: #e2e8f0;
  margin: 1.5rem 0;
}
</style>