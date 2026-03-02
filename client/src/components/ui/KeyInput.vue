<script setup lang="ts">
import { ref, computed } from 'vue'
import { EyeIcon, EyeSlashIcon } from '@heroicons/vue/24/outline'

const props = defineProps<{
  modelValue: string
  placeholder?: string
  state?: 'idle' | 'testing' | 'valid' | 'invalid'
  label?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  change: [value: string]
}>()

const showKey = ref(false)

const inputType = computed(() => (showKey.value ? 'text' : 'password'))

const borderClass = computed(() => ({
  idle: 'border-surface-600 focus-within:border-primary',
  testing: 'border-warning',
  valid: 'border-success',
  invalid: 'border-danger',
}[props.state ?? 'idle']))

const maskedDisplay = computed(() => {
  if (showKey.value || !props.modelValue) return props.modelValue
  if (props.modelValue.length > 8) {
    return '••••••••' + props.modelValue.slice(-4)
  }
  return '••••••••'
})

function handleInput(e: Event) {
  const val = (e.target as HTMLInputElement).value
  emit('update:modelValue', val)
  emit('change', val)
}
</script>

<template>
  <div>
    <label v-if="label" class="block text-xs font-medium text-muted mb-1.5 uppercase tracking-wider">
      {{ label }}
    </label>
    <div
      :class="[
        'flex items-center gap-2 px-3 py-2.5 rounded-md bg-surface-700 border transition-colors duration-120',
        borderClass,
      ]"
    >
      <input
        :type="inputType"
        :value="modelValue"
        :placeholder="placeholder || 'Enter API key...'"
        @input="handleInput"
        class="flex-1 bg-transparent text-sm font-mono text-text placeholder-muted outline-none min-w-0"
        autocomplete="off"
        autocorrect="off"
        spellcheck="false"
      />

      <!-- State indicator -->
      <div v-if="state === 'testing'" class="w-3 h-3 rounded-full border-2 border-warning border-t-transparent animate-spin flex-shrink-0" />
      <div v-else-if="state === 'valid'" class="text-success flex-shrink-0">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
      </div>
      <div v-else-if="state === 'invalid'" class="text-danger flex-shrink-0">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </div>

      <button
        type="button"
        @click="showKey = !showKey"
        class="text-muted hover:text-text transition-colors duration-120 flex-shrink-0"
      >
        <EyeSlashIcon v-if="showKey" class="w-4 h-4" />
        <EyeIcon v-else class="w-4 h-4" />
      </button>
    </div>
  </div>
</template>
