import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useSetupStore } from '@/stores/setup'

const routes: RouteRecordRaw[] = [
  // Root: the global guard handles redirect after checking setup state
  {
    path: '/',
    redirect: '/setup/welcome', // placeholder — guard will override
  },

  // Setup routes
  {
    path: '/setup',
    component: () => import('@/views/setup/WizardLayout.vue'),
    children: [
      { path: '', redirect: '/setup/welcome' },
      { path: 'welcome',      name: 'setup-welcome',       component: () => import('@/views/setup/StepWelcome.vue') },
      { path: 'system-check', name: 'setup-system-check',  component: () => import('@/views/setup/StepSystemCheck.vue') },
      { path: 'cloudflare',   name: 'setup-cloudflare',    component: () => import('@/views/setup/StepCloudflare.vue') },
      { path: 'github',       name: 'setup-github',        component: () => import('@/views/setup/StepGitHub.vue') },
      { path: 'providers',    name: 'setup-providers',     component: () => import('@/views/setup/StepProviders.vue') },
      { path: 'opencode',     name: 'setup-opencode',      component: () => import('@/views/setup/StepOpenCode.vue') },
      { path: 'nanoclaw',     name: 'setup-nanoclaw',      component: () => import('@/views/setup/StepNanoClaw.vue') },
      { path: 'complete',     name: 'setup-complete',      component: () => import('@/views/setup/StepComplete.vue') },
    ],
  },

  // App routes
  {
    path: '/app',
    component: () => import('@/components/layout/AppShell.vue'),
    children: [
      { path: '',          redirect: '/app/dashboard' },
      { path: 'dashboard', name: 'dashboard', component: () => import('@/views/DashboardView.vue') },
      { path: 'ide/:projectId?', name: 'ide', component: () => import('@/views/IDEView.vue') },
      { path: 'agent',     name: 'agent',     component: () => import('@/views/AgentView.vue') },
      { path: 'projects',  name: 'projects',  component: () => import('@/views/ProjectsView.vue') },
      { path: 'github',    name: 'github',    component: () => import('@/views/GitHubView.vue') },
      { path: 'settings',  name: 'settings',  component: () => import('@/views/SettingsView.vue') },
    ],
  },

  // GitHub OAuth callback handled server-side; if client sees it, send to dashboard
  { path: '/auth/github/callback', redirect: '/app/dashboard' },

  // SPA 404 — send to root (guard will handle)
  { path: '/:pathMatch(.*)*', redirect: '/setup/welcome' },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    return { top: 0 }
  },
})

// Single global guard — fetches setup state exactly once, then caches in store
router.beforeEach(async (to, _from, next) => {
  // Skip the guard for setup routes themselves and auth
  const isSetup = to.path.startsWith('/setup')
  const isApp   = to.path.startsWith('/app')

  if (!isSetup && !isApp) {
    // root, auth, 404 — just let the static redirect resolve
    return next()
  }

  const setupStore = useSetupStore()

  // Fetch once
  if (!setupStore.loaded) {
    await setupStore.fetchState()
  }

  const complete = setupStore.currentStep === 'complete'

  if (isApp && !complete) {
    // Tried to access app before setup is done
    return next('/setup/welcome')
  }

  if (isSetup && complete && to.name !== 'setup-complete') {
    // Setup already done — go to dashboard
    return next('/app/dashboard')
  }

  next()
})

export default router
