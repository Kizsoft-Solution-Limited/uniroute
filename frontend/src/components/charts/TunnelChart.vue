<template>
  <div class="w-full h-full bg-slate-800/30 rounded-lg p-6 border border-slate-700/50">
    <svg :width="props.width" :height="props.height" :viewBox="`0 0 ${props.width} ${props.height}`" class="w-full h-full" preserveAspectRatio="xMidYMid meet">
      <!-- Gradients and filters -->
      <defs>
        <linearGradient id="areaGradient" x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%" style="stop-color:#a855f7;stop-opacity:0.4" />
          <stop offset="50%" style="stop-color:#a855f7;stop-opacity:0.2" />
          <stop offset="100%" style="stop-color:#a855f7;stop-opacity:0" />
        </linearGradient>
        <linearGradient id="lineGradient" x1="0%" y1="0%" x2="100%" y2="0%">
          <stop offset="0%" style="stop-color:#c084fc" />
          <stop offset="50%" style="stop-color:#a855f7" />
          <stop offset="100%" style="stop-color:#9333ea" />
        </linearGradient>
        <filter id="glow">
          <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
          <feMerge>
            <feMergeNode in="coloredBlur"/>
            <feMergeNode in="SourceGraphic"/>
          </feMerge>
        </filter>
        <filter id="pointGlow">
          <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
          <feMerge>
            <feMergeNode in="coloredBlur"/>
            <feMergeNode in="SourceGraphic"/>
          </feMerge>
        </filter>
      </defs>
      
      <!-- Background grid -->
      <g v-for="(label, index) in yAxisLabels" :key="'grid-' + index" class="text-slate-600">
        <line
          :x1="padding"
          :y1="padding + ((props.height - padding * 2 - xAxisHeight) / (yAxisLabels.length - 1)) * index"
          :x2="props.width - padding"
          :y2="padding + ((props.height - padding * 2 - xAxisHeight) / (yAxisLabels.length - 1)) * index"
          stroke="currentColor"
          stroke-width="1"
          opacity="0.15"
        />
      </g>

      <!-- Y-axis labels -->
      <g v-for="(label, index) in yAxisLabels" :key="'label-' + index" class="text-xs text-slate-300">
        <text
          :x="padding - 10"
          :y="padding + ((props.height - padding * 2 - xAxisHeight) / (yAxisLabels.length - 1)) * index"
          text-anchor="end"
          dominant-baseline="middle"
          fill="currentColor"
          font-weight="500"
        >
          {{ label }}
        </text>
      </g>

      <!-- Area under curve (gradient fill) -->
      <path
        :d="areaPath"
        fill="url(#areaGradient)"
        class="transition-all duration-500 ease-out"
        opacity="0.8"
      />

      <!-- Main trend line -->
      <path
        :d="linePath"
        fill="none"
        stroke="url(#lineGradient)"
        stroke-width="3"
        stroke-linecap="round"
        stroke-linejoin="round"
        filter="url(#glow)"
        class="transition-all duration-500 ease-out"
      />

      <!-- Data points (circles) with hover effects -->
      <g v-for="(point, index) in chartPoints" :key="index">
        <circle
          :cx="point.x"
          :cy="point.y"
          r="5"
          fill="#a855f7"
          stroke="#ffffff"
          stroke-width="2"
          class="transition-all cursor-pointer hover:r-7"
          filter="url(#pointGlow)"
          opacity="0.9"
          @mouseenter="hoveredIndex = index"
          @mouseleave="hoveredIndex = null"
        >
          <title>{{ point.active }} active tunnels ({{ point.total }} total) at {{ formatTime(point.time) }}</title>
        </circle>
        
        <!-- Value label on hover or latest point -->
        <text
          v-if="hoveredIndex === index || index === chartPoints.length - 1"
          :x="point.x"
          :y="point.y - 15"
          text-anchor="middle"
          :fill="index === chartPoints.length - 1 ? '#a855f7' : '#c084fc'"
          font-size="13"
          font-weight="700"
          class="pointer-events-none transition-opacity"
        >
          {{ point.active }}
        </text>
        
        <!-- Active label for latest point -->
        <text
          v-if="index === chartPoints.length - 1 && point.active > 0"
          :x="point.x"
          :y="point.y - 30"
          text-anchor="middle"
          fill="#c084fc"
          font-size="10"
          font-weight="600"
          class="pointer-events-none"
        >
          active
        </text>
      </g>

      <!-- X-axis time labels - show every Nth label to avoid crowding -->
      <g v-for="(point, index) in chartPoints" :key="'xlabel-' + index">
        <text
          v-if="shouldShowXLabel(index)"
          :x="point.x"
          :y="props.height - padding - xAxisHeight + 20"
          text-anchor="middle"
          fill="#94a3b8"
          font-size="11"
          font-weight="500"
          :transform="point.labelRotation"
          class="pointer-events-none"
        >
          {{ formatTimeShort(point.time) }}
        </text>
      </g>

      <!-- Hover tooltip -->
      <g v-if="hoveredIndex !== null && chartPoints[hoveredIndex]">
        <rect
          :x="chartPoints[hoveredIndex].x - 50"
          :y="chartPoints[hoveredIndex].y - 50"
          width="100"
          height="35"
          rx="6"
          fill="#1e293b"
          stroke="#a855f7"
          stroke-width="1"
          opacity="0.95"
          class="pointer-events-none"
        />
        <text
          :x="chartPoints[hoveredIndex].x"
          :y="chartPoints[hoveredIndex].y - 35"
          text-anchor="middle"
          fill="#ffffff"
          font-size="11"
          font-weight="600"
          class="pointer-events-none"
        >
          {{ chartPoints[hoveredIndex].active }} active
        </text>
        <text
          :x="chartPoints[hoveredIndex].x"
          :y="chartPoints[hoveredIndex].y - 20"
          text-anchor="middle"
          fill="#c084fc"
          font-size="10"
          class="pointer-events-none"
        >
          {{ formatTime(chartPoints[hoveredIndex].time) }}
        </text>
      </g>
    </svg>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
  data: Array<{
    time: string
    active_tunnels: number
    total_tunnels: number
  }>
  width?: number
  height?: number
  hours?: number
}

const props = withDefaults(defineProps<Props>(), {
  width: 800,
  height: 400,
  hours: 24
})

const padding = 60
const xAxisHeight = 50
const hoveredIndex = ref<number | null>(null)

const maxValue = computed(() => {
  if (props.data.length === 0) return 10
  const max = Math.max(...props.data.map(d => Math.max(d.active_tunnels, d.total_tunnels)))
  return Math.max(1, Math.ceil(max * 1.2)) // Add 20% padding, minimum 1
})

// Generate evenly spaced Y-axis labels (0 to maxValue) without duplicates
const yAxisLabels = computed(() => {
  const max = maxValue.value
  const numLabels = 6 // Number of labels on Y-axis
  const labels: number[] = []
  
  // For small max values, use integer steps
  if (max <= 5) {
    // Simple case: use integers from 0 to max
    for (let i = max; i >= 0; i--) {
      labels.push(i)
    }
    return labels
  }
  
  // For larger values, generate evenly spaced labels
  const step = max / (numLabels - 1)
  for (let i = 0; i < numLabels; i++) {
    const value = Math.round(max - (step * i))
    // Avoid duplicates
    if (labels.length === 0 || labels[labels.length - 1] !== value) {
      labels.push(value)
    }
  }
  
  // Ensure 0 is included
  if (labels[labels.length - 1] !== 0) {
    labels.push(0)
  }
  
  return labels
})

const chartPoints = computed(() => {
  if (props.data.length === 0) return []
  
  const chartWidth = props.width - padding * 2
  const chartHeight = props.height - padding * 2 - xAxisHeight
  
  return props.data.map((point, index) => {
    const x = padding + (chartWidth / (props.data.length - 1 || 1)) * index
    const y = padding + chartHeight - (point.active_tunnels / maxValue.value) * chartHeight
    
    // Rotate labels if there are many data points to prevent overlap
    const needsRotation = props.data.length > 12
    const rotationAngle = needsRotation ? -45 : 0
    const labelX = x
    const labelY = props.height - padding - xAxisHeight + 20
    const labelRotation = needsRotation 
      ? `rotate(${rotationAngle} ${labelX} ${labelY})`
      : ''
    
    return {
      x,
      y,
      active: point.active_tunnels,
      total: point.total_tunnels,
      time: point.time,
      labelRotation
    }
  })
})

const linePath = computed(() => {
  if (chartPoints.value.length === 0) return ''
  
  const points = chartPoints.value
  if (points.length === 1) {
    return `M ${points[0].x} ${points[0].y} L ${points[0].x} ${points[0].y}`
  }
  
  // Smooth curve using cubic bezier for trending effect
  let path = `M ${points[0].x} ${points[0].y}`
  
  for (let i = 1; i < points.length; i++) {
    const prev = points[i - 1]
    const curr = points[i]
    
    // Calculate control points for smooth curves
    const cp1x = prev.x + (curr.x - prev.x) / 3
    const cp1y = prev.y
    const cp2x = prev.x + (curr.x - prev.x) * 2 / 3
    const cp2y = curr.y
    
    path += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${curr.x} ${curr.y}`
  }
  
  return path
})

const areaPath = computed(() => {
  if (chartPoints.value.length === 0) return ''
  
  const points = chartPoints.value
  const bottomY = props.height - padding - xAxisHeight
  
  if (points.length === 1) {
    return `M ${points[0].x} ${bottomY} L ${points[0].x} ${points[0].y} L ${points[0].x} ${bottomY} Z`
  }
  
  // Create smooth area path matching the line
  let path = `M ${points[0].x} ${bottomY}`
  path += ` L ${points[0].x} ${points[0].y}`
  
  for (let i = 1; i < points.length; i++) {
    const prev = points[i - 1]
    const curr = points[i]
    
    // Use same control points as the line for consistency
    const cp1x = prev.x + (curr.x - prev.x) / 3
    const cp1y = prev.y
    const cp2x = prev.x + (curr.x - prev.x) * 2 / 3
    const cp2y = curr.y
    
    path += ` C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${curr.x} ${curr.y}`
  }
  
  const lastPoint = points[points.length - 1]
  path += ` L ${lastPoint.x} ${bottomY} Z`
  
  return path
})

function formatTime(timeStr: string): string {
  const date = new Date(timeStr)
  return date.toLocaleTimeString('en-US', { 
    hour: '2-digit', 
    minute: '2-digit',
    hour12: false 
  })
}

// Determine if we should show X-axis label for this point (to avoid crowding)
function shouldShowXLabel(index: number): boolean {
  const totalPoints = chartPoints.value.length
  
  // Always show first and last labels
  if (index === 0 || index === totalPoints - 1) {
    return true
  }
  
  // For many points, show every Nth label
  if (totalPoints > 12) {
    // Show every 3rd label for crowded charts
    return index % 3 === 0
  } else if (totalPoints > 6) {
    // Show every 2nd label for medium charts
    return index % 2 === 0
  }
  
  // Show all labels for small charts
  return true
}

function formatTimeShort(timeStr: string): string {
  const date = new Date(timeStr)
  
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const time = date.toLocaleTimeString('en-US', { 
    hour: '2-digit', 
    minute: '2-digit', 
    hour12: false 
  })
  
  // For 6h view: show time only (HH:MM) - time-series data over 6 hours
  if (props.hours <= 6) {
    return time
  }
  
  // For 7d view: show date only (MM/DD) since we're showing one data point per day
  if (props.hours >= 168) {
    return `${month}/${day}`
  }
  
  // For 24h view: show time only (HH:MM) - time-series data over 24 hours
  return time
}
</script>
