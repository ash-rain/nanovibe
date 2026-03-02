<script setup lang="ts">
import { computed, getCurrentInstance } from 'vue'

const props = defineProps<{
  data: number[]
  color?: string
  height?: number
  smooth?: boolean
}>()

const instance = getCurrentInstance()
const uid = instance ? instance.uid : Math.random().toString(36).slice(2)

const heightVal = computed(() => props.height ?? 32)
const color = computed(() => props.color ?? 'var(--color-primary-glow)')
const width = 120

const viewBoxStr = computed(() => `0 0 ${width} ${heightVal.value}`)

const path = computed(() => {
  const data = props.data
  if (!data || data.length < 2) return ''

  const min = Math.min(...data)
  const max = Math.max(...data)
  const range = max - min || 1

  const points = data.map((v, i) => {
    const x = (i / (data.length - 1)) * width
    const y = heightVal.value - ((v - min) / range) * (heightVal.value * 0.85) - 2
    return [x, y] as [number, number]
  })

  if (props.smooth === false) {
    return points.map((p, i) => `${i === 0 ? 'M' : 'L'}${p[0]},${p[1]}`).join(' ')
  }

  if (points.length < 2) return ''

  let d = `M${points[0][0]},${points[0][1]}`
  for (let i = 1; i < points.length; i++) {
    const prev = points[i - 1]
    const curr = points[i]
    const cpX = (prev[0] + curr[0]) / 2
    d += ` C${cpX},${prev[1]} ${cpX},${curr[1]} ${curr[0]},${curr[1]}`
  }
  return d
})

const areaPath = computed(() => {
  if (!path.value) return ''
  const data = props.data
  if (!data || data.length < 2) return ''

  const lastX = width
  const firstX = 0

  return `${path.value} L${lastX},${heightVal.value} L${firstX},${heightVal.value} Z`
})

const gradId = `grad-${uid}`
</script>

<template>
  <svg
    :width="width"
    :height="heightVal"
    :viewBox="viewBoxStr"
    :style="`width: 100%; height: ${heightVal}px; display: block;`"
    preserveAspectRatio="none"
    overflow="visible"
  >
    <defs>
      <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" :stop-color="color" stop-opacity="0.2" />
        <stop offset="100%" :stop-color="color" stop-opacity="0" />
      </linearGradient>
    </defs>

    <!-- Area fill -->
    <path
      v-if="areaPath"
      :d="areaPath"
      :fill="`url(#${gradId})`"
    />

    <!-- Line -->
    <path
      v-if="path"
      :d="path"
      fill="none"
      :stroke="color"
      stroke-width="1.5"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
  </svg>
</template>
