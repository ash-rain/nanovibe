<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import QRCode from 'qrcode'

const props = defineProps<{
  url: string
  size?: number
}>()

const canvasRef = ref<HTMLCanvasElement | null>(null)

async function generateQR() {
  if (!canvasRef.value || !props.url) return
  try {
    await QRCode.toCanvas(canvasRef.value, props.url, {
      width: props.size ?? 160,
      color: {
        dark: '#ffffff',
        light: '#0a0a14',
      },
      margin: 2,
      errorCorrectionLevel: 'M',
    })
  } catch (err) {
    console.error('QR generation failed:', err)
  }
}

onMounted(generateQR)
watch(() => props.url, generateQR)
</script>

<template>
  <div class="inline-flex items-center justify-center p-2 rounded-md bg-surface-900">
    <canvas ref="canvasRef" class="rounded" />
  </div>
</template>
