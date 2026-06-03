<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="modal-overlay" @click.self="close">
        <div class="modal-container">
          <div class="modal-header">
            <h3>个人资料</h3>
            <button class="btn-close" @click="close">
              <Icon icon="mdi:close" />
            </button>
          </div>
          <div class="modal-body">
            <div class="profile-info">
              <div class="avatar-huge">
                <Icon icon="mdi:user-circle" />
              </div>
              <div class="info-list">
                <div class="info-item">
                  <span class="info-label">用户名</span>
                  <span class="info-value">{{ userStore.user?.username }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">邮箱</span>
                  <span class="info-value">{{ userStore.user?.email }}</span>
                </div>
                <div class="info-item">
                  <span class="info-label">注册时间</span>
                  <span class="info-value">{{ formatDate(userStore.user?.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { Icon } from '@iconify/vue'
import { useUserStore } from '../../stores/user'

const props = defineProps({
  show: Boolean
})

const emit = defineEmits(['close'])

const userStore = useUserStore()

const close = () => {
  emit('close')
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  padding: 1rem;
}

.modal-container {
  background: var(--card-bg, white);
  border-radius: 16px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1.2rem 1.5rem;
  border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.08));
}

.modal-header h3 {
  margin: 0;
  font-size: 1.2rem;
  color: var(--text-primary, #212529);
}

.btn-close {
  background: none;
  border: none;
  color: var(--text-secondary, #6c757d);
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.btn-close:hover {
  background: var(--btn-secondary-bg, #f8f9fa);
  color: var(--text-primary, #212529);
}

.modal-body {
  padding: 1.5rem;
}

.profile-info {
  text-align: center;
  padding: 1rem 0;
}

.avatar-huge {
  width: 100px;
  height: 100px;
  margin: 0 auto 1.5rem;
  background: linear-gradient(135deg, #4361ee 0%, #7209b7 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 3rem;
  color: white;
}

.info-list {
  text-align: left;
}

.info-item {
  display: flex;
  justify-content: space-between;
  padding: 0.8rem 0;
  border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.08));
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  color: var(--text-secondary, #6c757d);
  font-size: 0.9rem;
}

.info-value {
  color: var(--text-primary, #212529);
  font-weight: 500;
  font-size: 0.9rem;
}
</style>
