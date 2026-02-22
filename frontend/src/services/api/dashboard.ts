import { apiClient } from './client'
import { analyticsApi } from './analytics'
import { apiKeysApi } from './apikeys'
import { tunnelsApi } from './tunnels'

export async function getAdminOverviewStats(): Promise<{
  total_requests: number
  request_growth: number
  active_tunnels: number
  total_tunnels: number
}> {
  const now = new Date()
  const currentStart = new Date(now.getFullYear(), now.getMonth(), 1).toISOString()
  const currentEnd = now.toISOString()
  const prevStart = new Date(now.getFullYear(), now.getMonth() - 1, 1).toISOString()
  const prevEnd = new Date(now.getFullYear(), now.getMonth(), 0).toISOString()

  const [currentStats, previousStats, counts] = await Promise.all([
    analyticsApi.getUsageStatsAdmin(currentStart, currentEnd),
    analyticsApi.getUsageStatsAdmin(prevStart, prevEnd),
    tunnelsApi.getCountsAdmin()
  ])

  const requestGrowth =
    previousStats.total_requests > 0
      ? Math.round(
          ((currentStats.total_requests - previousStats.total_requests) / previousStats.total_requests) * 100
        )
      : 0

  return {
    total_requests: currentStats.total_requests,
    request_growth: requestGrowth,
    active_tunnels: counts.active,
    total_tunnels: counts.total
  }
}

export interface DashboardStats {
  total_requests: number
  request_growth: number // Percentage change from previous period
  active_tunnels: number
  total_tunnels: number
  api_keys: number
  active_api_keys: number
  total_cost: number
  provider_usage: Array<{
    name: string
    icon: string
    percentage: number
    requests: number
    cost: number
  }>
  recent_activity: Array<{
    type: 'api_key' | 'tunnel' | 'request'
    title: string
    time: string
    icon: string
  }>
}

export const dashboardApi = {
  /**
   * Get dashboard overview statistics
   */
  async getOverview(): Promise<DashboardStats> {
    // Fetch data from multiple endpoints
    const [usageStats, apiKeys, tunnels, recentRequests] = await Promise.all([
      // Get usage stats for current month and previous month
      Promise.all([
        analyticsApi.getUsageStats(
          new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString(),
          new Date().toISOString()
        ),
        analyticsApi.getUsageStats(
          new Date(new Date().getFullYear(), new Date().getMonth() - 1, 1).toISOString(),
          new Date(new Date().getFullYear(), new Date().getMonth(), 0).toISOString()
        )
      ]),
      apiKeysApi.list(),
      tunnelsApi.list(),
      analyticsApi.getRequests(10, 0)
    ])

    const [currentStats, previousStats] = usageStats

    // Calculate growth percentage
    const requestGrowth = previousStats.total_requests > 0
      ? Math.round(((currentStats.total_requests - previousStats.total_requests) / previousStats.total_requests) * 100)
      : 0

    // Count active tunnels
    const activeTunnels = tunnels.tunnels.filter(t => t.status === 'active').length

    // Count active API keys
    const activeApiKeys = apiKeys.keys.filter(k => k.is_active).length

    // Calculate provider usage
    const totalRequests = currentStats.total_requests || 1 // Avoid division by zero
    const providerUsage = Object.entries(currentStats.requests_by_provider || {}).map(([name, requests]) => {
      const percentage = Math.round((requests / totalRequests) * 100)
      const cost = currentStats.cost_by_provider?.[name] || 0
      
      // Map provider names to icons
      const iconMap: Record<string, string> = {
        openai: '🤖',
        anthropic: '🧠',
        google: '🔍',
        'google-ai': '🔍',
        gemini: '💎',
        'meta-llama': '🦙',
        llama: '🦙',
        mistral: '🌊'
      }
      
      return {
        name: name.charAt(0).toUpperCase() + name.slice(1),
        icon: iconMap[name.toLowerCase()] || '🔧',
        percentage,
        requests: requests as number,
        cost
      }
    }).sort((a, b) => b.percentage - a.percentage)

    // Generate recent activity from recent requests
    const recentActivity = recentRequests.requests.slice(0, 5).map(req => {
      const timeAgo = getTimeAgo(new Date(req.created_at))
      return {
        type: 'request' as const,
        title: `${req.provider} request completed`,
        time: timeAgo,
        icon: '📊'
      }
    })

    // Add API key and tunnel activities if available
    if (apiKeys.keys.length > 0) {
      const latestKey = apiKeys.keys.sort((a, b) => 
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      )[0]
      recentActivity.unshift({
        type: 'api_key',
        title: `API key "${latestKey.name}" created`,
        time: getTimeAgo(new Date(latestKey.created_at)),
        icon: '🔑'
      })
    }

    if (tunnels.tunnels.length > 0) {
      const latestTunnel = tunnels.tunnels.sort((a, b) => 
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      )[0]
      if (latestTunnel.status === 'active') {
        recentActivity.unshift({
          type: 'tunnel',
          title: `Tunnel "${latestTunnel.subdomain}" connected`,
          time: getTimeAgo(new Date(latestTunnel.created_at)),
          icon: '🌐'
        })
      }
    }

    return {
      total_requests: currentStats.total_requests,
      request_growth: requestGrowth,
      active_tunnels: activeTunnels,
      total_tunnels: tunnels.tunnels.length,
      api_keys: apiKeys.keys.length,
      active_api_keys: activeApiKeys,
      total_cost: currentStats.total_cost,
      provider_usage: providerUsage,
      recent_activity: recentActivity.slice(0, 5) // Limit to 5 most recent
    }
  }
}

function getTimeAgo(date: Date): string {
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`
  if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
  if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
  return date.toLocaleDateString()
}

