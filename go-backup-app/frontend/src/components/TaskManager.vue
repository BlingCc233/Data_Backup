<template>
  <div class="tasks-view">
    <div class="card neo-card">
      <div class="card-label">Tasks</div>
      <h3>定时 / 实时备份</h3>
      <p class="description">支持 cron 定时任务与文件变更实时备份（默认增量）。</p>

      <div class="action-bar">
        <button class="primary brutal-btn" @click="openCreateModal">新建任务</button>
        <button class="brutal-btn" @click="refresh">刷新</button>
      </div>
    </div>

    <div class="card neo-card" v-if="tasks.length === 0">
      <p class="empty-dir-msg">暂无任务</p>
    </div>

    <div class="card neo-card" v-for="t in tasks" :key="t.id">
      <div class="task-row">
        <div class="task-main">
          <div class="task-title">
            <span class="font-mono">{{ t.id }}</span>
            <strong>{{ t.name }}</strong>
            <span class="task-type">{{ t.type }}</span>
          </div>
          <div class="task-sub font-mono">
            <div v-if="t.type === 'schedule'">cron: {{ t.config?.cronExpr }}</div>
            <div v-if="t.type === 'watch'">watch: {{ (t.config?.watchPaths || []).join(', ') }}</div>
            <div>dest: {{ t.config?.destinationDir }}</div>
            <div v-if="t.config?.lastBackupPath">last: {{ t.config?.lastBackupPath }}</div>
          </div>
        </div>

        <div class="task-actions">
          <label class="switch-inline">
            <span>启用</span>
            <input type="checkbox" v-model="t.enabled" @change="toggleEnabled(t)" />
          </label>
          <button class="brutal-btn" @click="runNow(t)">运行</button>
          <button class="brutal-btn" @click="openEditModal(t)">编辑</button>
          <button class="secondary brutal-btn" @click="removeTask(t)">删除</button>
        </div>
      </div>
    </div>

    <div v-if="modalVisible" class="modal-overlay">
      <div class="modal-card">
        <h3>{{ editingId ? '编辑任务' : '新建任务' }}</h3>

        <div class="input-group">
          <label>名称</label>
          <input v-model="form.name" type="text" placeholder="task name" />
        </div>

        <div class="input-group">
          <label>类型</label>
          <select v-model="form.type">
            <option value="schedule">schedule (cron)</option>
            <option value="watch">watch (实时)</option>
          </select>
        </div>

        <div class="input-group" v-if="form.type === 'schedule'">
          <label>Cron 表达式</label>
          <input v-model="form.cronExpr" type="text" placeholder="@every 1h 或 0 2 * * *" />
        </div>

        <div v-if="form.type === 'watch'">
          <div class="input-group">
            <label>Watch Paths (可选，默认使用 SourcePaths)</label>
            <textarea v-model="form.watchPathsText" placeholder="/path/to/watch"></textarea>
          </div>
          <div class="input-group">
            <label>Debounce (ms)</label>
            <input v-model.number="form.watchDebounceMs" type="number" min="50" />
          </div>
        </div>

        <div class="card-divider"></div>

        <div class="input-group">
          <label>Source Paths</label>
          <textarea v-model="form.sourcePathsText" placeholder="/path/to/source"></textarea>
          <div class="inline-buttons">
            <button class="brutal-btn" @click="addSourceDir">添加文件夹</button>
            <button class="brutal-btn" @click="addSourceFiles">添加文件</button>
          </div>
        </div>

        <div class="input-group">
          <label>Destination Dir</label>
          <input v-model="form.destinationDir" readonly type="text" placeholder="选择目录..." />
          <button class="brutal-btn" @click="browseDest">浏览</button>
        </div>

        <div class="card-divider"></div>

        <div class="input-group switch-inline">
          <span>增量备份</span>
          <input type="checkbox" v-model="form.incremental" />
        </div>
        <div class="input-group switch-inline">
          <span>启用压缩</span>
          <input type="checkbox" v-model="form.useCompression" />
        </div>
        <div class="input-group switch-inline">
          <span>启用加密</span>
          <input type="checkbox" v-model="form.useEncryption" />
        </div>
        <div v-if="form.useEncryption" class="encryption-options">
          <div class="input-group">
            <label>密码</label>
            <input v-model="form.password" type="password" placeholder="PASSWORD" />
          </div>
          <div class="input-group">
            <label>算法</label>
            <select v-model.number="form.algorithm">
              <option :value="1">AES-256</option>
              <option :value="2">ChaCha20</option>
            </select>
          </div>
        </div>

        <div class="card-divider"></div>

        <h4>高级筛选</h4>
        <div class="filter-grid">
          <div class="filter-group">
            <label>包含名称</label>
            <textarea v-model="form.filters.includeNames" placeholder="*.log"></textarea>
          </div>
          <div class="filter-group">
            <label>排除名称</label>
            <textarea v-model="form.filters.excludeNames" placeholder="*.tmp"></textarea>
          </div>
          <div class="filter-group">
            <label>包含路径</label>
            <textarea v-model="form.filters.includePaths" placeholder="/path/to/include"></textarea>
          </div>
          <div class="filter-group">
            <label>排除路径</label>
            <textarea v-model="form.filters.excludePaths" placeholder="/path/to/exclude"></textarea>
          </div>
          <div class="filter-group">
            <label>最小大小</label>
            <div class="size-input">
              <input type="number" v-model.number="form.filters.minSizeValue" min="0" />
              <select v-model="form.filters.minSizeUnit">
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
              <input type="number" v-model.number="form.filters.maxSizeValue" min="0" />
              <select v-model="form.filters.maxSizeUnit">
                <option>Bytes</option>
                <option>KB</option>
                <option>MB</option>
                <option>GB</option>
              </select>
            </div>
          </div>
          <div class="filter-group">
            <label>修改时间晚于</label>
            <input type="datetime-local" v-model="form.filters.newerThan" />
          </div>
          <div class="filter-group">
            <label>修改时间早于</label>
            <input type="datetime-local" v-model="form.filters.olderThan" />
          </div>
        </div>

        <div class="action-bar">
          <button class="brutal-btn" @click="closeModal">取消</button>
          <button class="primary brutal-btn" @click="saveTask">{{ editingId ? '保存' : '创建' }}</button>
        </div>

        <p class="font-mono" v-if="errorMessage">{{ errorMessage }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue';
import { CreateTask, DeleteTask, GetTasks, RunTaskNow, SelectDirectory, SelectFiles, UpdateTask } from '../../wailsjs/go/main/App';

const tasks = ref([]);
const modalVisible = ref(false);
const editingId = ref('');
const errorMessage = ref('');

const form = reactive({
  name: '',
  type: 'schedule',
  cronExpr: '@every 1h',
  watchPathsText: '',
  watchDebounceMs: 500,
  sourcePathsText: '',
  destinationDir: '',
  incremental: true,
  useCompression: true,
  useEncryption: false,
  algorithm: 1,
  password: '',
  filters: {
    includeNames: '',
    excludeNames: '',
    includePaths: '',
    excludePaths: '',
    minSizeValue: 0,
    minSizeUnit: 'Bytes',
    maxSizeValue: 0,
    maxSizeUnit: 'Bytes',
    newerThan: null,
    olderThan: null,
  },
});

function normalizeLines(text) {
  return text.split('\n').map(s => s.trim()).filter(Boolean);
}

function convertToBytes(value, unit) {
  const multipliers = { Bytes: 1, KB: 1024, MB: 1024 ** 2, GB: 1024 ** 3 };
  return (value || 0) * (multipliers[unit] || 1);
}

function buildFilterConfig() {
  const minSize = convertToBytes(form.filters.minSizeValue, form.filters.minSizeUnit);
  let maxSize = convertToBytes(form.filters.maxSizeValue, form.filters.maxSizeUnit);
  if (maxSize <= 0) maxSize = -1;
  return {
    includePaths: normalizeLines(form.filters.includePaths),
    excludePaths: normalizeLines(form.filters.excludePaths),
    includeNames: normalizeLines(form.filters.includeNames),
    excludeNames: normalizeLines(form.filters.excludeNames),
    newerThan: form.filters.newerThan ? new Date(form.filters.newerThan).toISOString() : null,
    olderThan: form.filters.olderThan ? new Date(form.filters.olderThan).toISOString() : null,
    minSize,
    maxSize,
  };
}

async function refresh() {
  tasks.value = await GetTasks();
}

function resetForm() {
  editingId.value = '';
  errorMessage.value = '';
  form.name = '';
  form.type = 'schedule';
  form.cronExpr = '@every 1h';
  form.watchPathsText = '';
  form.watchDebounceMs = 500;
  form.sourcePathsText = '';
  form.destinationDir = '';
  form.incremental = true;
  form.useCompression = true;
  form.useEncryption = false;
  form.algorithm = 1;
  form.password = '';
  Object.assign(form.filters, {
    includeNames: '',
    excludeNames: '',
    includePaths: '',
    excludePaths: '',
    minSizeValue: 0,
    minSizeUnit: 'Bytes',
    maxSizeValue: 0,
    maxSizeUnit: 'Bytes',
    newerThan: null,
    olderThan: null,
  });
}

function openCreateModal() {
  resetForm();
  modalVisible.value = true;
}

function openEditModal(task) {
  resetForm();
  editingId.value = task.id;
  form.name = task.name;
  form.type = task.type;
  form.cronExpr = task.config?.cronExpr || '@every 1h';
  form.watchDebounceMs = task.config?.watchDebounceMs || 500;
  form.watchPathsText = (task.config?.watchPaths || []).join('\n');
  form.sourcePathsText = (task.config?.sourcePaths || []).join('\n');
  form.destinationDir = task.config?.destinationDir || '';
  form.incremental = !!task.config?.incremental;
  form.useCompression = !!task.config?.useCompression;
  form.useEncryption = !!task.config?.useEncryption;
  form.algorithm = task.config?.algorithm || 1;
  form.password = task.config?.password || '';

  const f = task.config?.filters || {};
  form.filters.includeNames = (f.includeNames || []).join('\n');
  form.filters.excludeNames = (f.excludeNames || []).join('\n');
  form.filters.includePaths = (f.includePaths || []).join('\n');
  form.filters.excludePaths = (f.excludePaths || []).join('\n');
  form.filters.minSizeValue = 0;
  form.filters.minSizeUnit = 'Bytes';
  form.filters.maxSizeValue = 0;
  form.filters.maxSizeUnit = 'Bytes';
  form.filters.newerThan = null;
  form.filters.olderThan = null;

  modalVisible.value = true;
}

function closeModal() {
  modalVisible.value = false;
}

async function browseDest() {
  const dir = await SelectDirectory();
  if (dir) form.destinationDir = dir;
}

async function addSourceDir() {
  const paths = await SelectFiles(true);
  if (paths && paths.length > 0) {
    form.sourcePathsText = [form.sourcePathsText, ...paths].filter(Boolean).join('\n');
  }
}

async function addSourceFiles() {
  const paths = await SelectFiles(false);
  if (paths && paths.length > 0) {
    form.sourcePathsText = [form.sourcePathsText, ...paths].filter(Boolean).join('\n');
  }
}

async function saveTask() {
  errorMessage.value = '';
  try {
    const sourcePaths = normalizeLines(form.sourcePathsText);
    const watchPaths = normalizeLines(form.watchPathsText);
    const payload = {
      id: editingId.value || '',
      name: form.name,
      type: form.type,
      enabled: true,
      config: {
        sourcePaths,
        destinationDir: form.destinationDir,
        filters: buildFilterConfig(),
        useCompression: form.useCompression,
        useEncryption: form.useEncryption,
        algorithm: form.algorithm,
        password: form.password,
        incremental: form.incremental,
        cronExpr: form.cronExpr,
        watchPaths: watchPaths.length > 0 ? watchPaths : sourcePaths,
        watchDebounceMs: form.watchDebounceMs,
      },
    };

    if (!payload.name.trim()) throw new Error('任务名称不能为空');
    if (!payload.config.destinationDir) throw new Error('请选择目标目录');
    if (payload.config.sourcePaths.length === 0) throw new Error('请填写 SourcePaths');
    if (payload.type === 'schedule' && !payload.config.cronExpr) throw new Error('Cron 表达式不能为空');
    if (payload.type === 'watch' && payload.config.watchPaths.length === 0) throw new Error('WatchPaths 不能为空');

    if (editingId.value) {
      await UpdateTask(payload);
    } else {
      await CreateTask(payload);
    }
    await refresh();
    closeModal();
  } catch (e) {
    errorMessage.value = `${e}`;
  }
}

async function removeTask(task) {
  await DeleteTask(task.id);
  await refresh();
}

async function runNow(task) {
  await RunTaskNow(task.id);
}

async function toggleEnabled(task) {
  const payload = {
    ...task,
    enabled: !!task.enabled,
  };
  await UpdateTask(payload);
  await refresh();
}

onMounted(refresh);
</script>

<style scoped>
.tasks-view {
  max-width: 1100px;
  margin: 0 auto;
}

.task-row {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: flex-start;
  justify-content: space-between;
}

.task-main {
  flex: 1 1 420px;
  min-width: 0;
}

.task-title {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.6rem;
}

.task-type {
  font-family: var(--font-mono);
  background: #eee;
  border: 2px solid #000;
  padding: 0 6px;
}

.task-sub > div {
  margin-top: 4px;
}

.task-actions {
  display: flex;
  flex: 0 0 auto;
  flex-wrap: wrap;
  gap: 0.6rem;
  align-items: center;
  justify-content: flex-end;
}

.action-bar {
  display: flex;
  gap: 0.8rem;
  flex-wrap: wrap;
  align-items: center;
  margin-top: 1rem;
}

.inline-buttons {
  display: flex;
  gap: 0.6rem;
  flex-wrap: wrap;
  margin-top: 0.6rem;
}

.input-group {
  margin-top: 0.8rem;
}

.input-group label {
  display: block;
  font-weight: 800;
  margin-bottom: 6px;
  text-transform: uppercase;
  font-size: 0.85rem;
}

.switch-inline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.encryption-options {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

textarea {
  min-height: 90px;
  resize: vertical;
}

/* Make the task modal usable in small windows */
.modal-overlay {
  align-items: flex-start;
  padding: 2rem 0;
  overflow-y: auto;
}

.modal-card {
  width: min(980px, 92vw);
  max-height: calc(100vh - 4rem);
  overflow-y: auto;
  background: white;
  border: 4px solid #000;
  box-shadow: 10px 10px 0 #000;
  padding: 1.25rem;
  box-sizing: border-box;
}

@media (max-width: 900px) {
  .filter-grid {
    grid-template-columns: 1fr;
  }
  .encryption-options {
    grid-template-columns: 1fr;
  }
}
</style>
