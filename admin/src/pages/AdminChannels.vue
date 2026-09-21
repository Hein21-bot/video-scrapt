<template>
  <div class="admin-channels">
    <h1 class="page-title">Channels</h1>
    <p class="page-sub">Main tabs shown on the user site. Built-in channels (Chinese AV, Myanmar, Muskuduu) can be renamed but not removed — they're wired to real scraper/upload code.</p>

    <form class="add-card" @submit.prevent="add">
      <input v-model="newLabel" class="add-input" placeholder="New channel name (e.g. Korean Drama)" />
      <button type="submit" class="add-btn" :disabled="!newLabel.trim() || adding">{{ adding ? '…' : '+ Add' }}</button>
    </form>
    <p v-if="err" class="err-msg">{{ err }}</p>

    <div class="table-wrap">
      <table class="table" v-if="channels.length">
        <thead>
          <tr><th>Label</th><th>Key</th><th>Videos</th><th></th></tr>
        </thead>
        <tbody>
          <tr v-for="ch in channels" :key="ch.key">
            <td>
              <input
                v-if="editingKey === ch.key" v-model="editLabel" class="edit-input"
                @keydown.enter="saveEdit(ch)" @keydown.escape="cancelEdit"
              />
              <span v-else>{{ ch.label }}</span>
              <span v-if="ch.builtin" class="builtin-badge">built-in</span>
            </td>
            <td class="dim">{{ ch.key }}</td>
            <td class="dim">{{ ch.count }}</td>
            <td class="actions">
              <template v-if="editingKey === ch.key">
                <button class="mini-btn" @click="saveEdit(ch)">Save</button>
                <button class="mini-btn ghost" @click="cancelEdit">Cancel</button>
              </template>
              <template v-else>
                <button class="mini-btn ghost" @click="startEdit(ch)">Rename</button>
                <button v-if="!ch.builtin" class="mini-btn danger" @click="remove(ch)">Delete</button>
              </template>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-else-if="loading" class="center"><div class="spinner" /></div>
      <div v-else class="center dim">No channels yet.</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { authHeader } from '../stores/auth'

interface Channel { id: string; key: string; label: string; count: number; builtin: boolean }

const channels = ref<Channel[]>([])
const loading  = ref(false)
const newLabel = ref('')
const adding   = ref(false)
const err      = ref('')

const editingKey = ref('')
const editLabel  = ref('')

async function load() {
  loading.value = true
  try {
    const res = await fetch('/api/admin/channels', { headers: authHeader() })
    channels.value = res.ok ? await res.json() : []
  } finally {
    loading.value = false
  }
}

async function add() {
  const label = newLabel.value.trim()
  if (!label) return
  adding.value = true
  err.value = ''
  try {
    const res  = await fetch('/api/admin/channels', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify({ label }),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error ?? 'Failed to add')
    newLabel.value = ''
    await load()
  } catch (e) {
    err.value = e instanceof Error ? e.message : 'Failed to add'
  } finally {
    adding.value = false
  }
}

function startEdit(ch: Channel) {
  editingKey.value = ch.key
  editLabel.value = ch.label
}
function cancelEdit() {
  editingKey.value = ''
}

async function saveEdit(ch: Channel) {
  const label = editLabel.value.trim()
  if (!label || label === ch.label) { cancelEdit(); return }
  err.value = ''
  try {
    const res  = await fetch(`/api/admin/channels/${ch.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json', ...authHeader() },
      body: JSON.stringify({ label }),
    })
    const data = await res.json()
    if (!res.ok) throw new Error(data.error ?? 'Failed to rename')
    cancelEdit()
    await load()
  } catch (e) {
    err.value = e instanceof Error ? e.message : 'Failed to rename'
  }
}

async function remove(ch: Channel) {
  if (!confirm(`Delete "${ch.label}"?`)) return
  const res  = await fetch(`/api/admin/channels/${ch.id}`, { method: 'DELETE', headers: authHeader() })
  const data = await res.json().catch(() => ({}))
  if (res.ok) {
    channels.value = channels.value.filter(x => x.key !== ch.key)
  } else {
    err.value = data.error ?? 'Failed to delete'
  }
}

onMounted(load)
</script>

<style scoped>
.admin-channels { display: flex; flex-direction: column; gap: 16px; }
.page-title { margin: 0; font-size: 1.4rem; font-weight: 700; color: #f1f5f9; }
.page-sub { margin: -10px 0 0; font-size: .85rem; color: #64748b; }

.add-card { display: flex; gap: 10px; }
.add-input {
  flex: 1; max-width: 320px; background: #1e293b; border: 1px solid #334155;
  border-radius: 7px; padding: 8px 12px; color: #f1f5f9; font-size: .85rem; outline: none;
  transition: border-color .15s;
}
.add-input:focus { border-color: #6366f1; }
.add-btn {
  background: #6366f1; border: none; border-radius: 7px; color: #fff;
  font-size: .85rem; font-weight: 600; padding: 8px 20px; cursor: pointer; transition: background .15s;
}
.add-btn:hover:not(:disabled) { background: #4f46e5; }
.add-btn:disabled { opacity: .6; cursor: not-allowed; }
.err-msg { margin: 0; font-size: .82rem; color: #f87171; }

.table-wrap { background: #1e293b; border: 1px solid #334155; border-radius: 10px; overflow: auto; }
.table { width: 100%; border-collapse: collapse; font-size: .83rem; }
.table th {
  padding: 10px 14px; text-align: left; font-size: .72rem;
  color: #64748b; text-transform: uppercase; letter-spacing: .05em;
  border-bottom: 1px solid #334155; white-space: nowrap;
}
.table td { padding: 10px 14px; border-bottom: 1px solid #1e293b; vertical-align: middle; color: #cbd5e1; }
.table tr:last-child td { border-bottom: none; }
.table tr:hover td { background: rgba(99,102,241,.05); }
.dim { color: #64748b; }
.builtin-badge { margin-left: 8px; font-size: .68rem; text-transform: uppercase; letter-spacing: .04em; color: #64748b; background: #0f172a; border: 1px solid #334155; border-radius: 4px; padding: 1px 6px; }

.edit-input {
  background: #0f172a; border: 1px solid #6366f1; border-radius: 5px;
  padding: 5px 8px; color: #f1f5f9; font-size: .83rem; outline: none; width: 160px;
}

.actions { display: flex; gap: 6px; white-space: nowrap; }
.mini-btn {
  background: transparent; border: 1px solid #334155; border-radius: 5px;
  color: #cbd5e1; cursor: pointer; padding: 5px 10px; font-size: .78rem; transition: all .15s;
}
.mini-btn:hover { border-color: #6366f1; color: #f1f5f9; }
.mini-btn.ghost { color: #94a3b8; }
.mini-btn.danger:hover { border-color: #f87171; color: #f87171; background: rgba(248,113,113,.08); }

.center { display: flex; justify-content: center; padding: 40px; }
.spinner { width: 28px; height: 28px; border: 3px solid rgba(99,102,241,.3); border-top-color: #6366f1; border-radius: 50%; animation: spin .8s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
</style>
