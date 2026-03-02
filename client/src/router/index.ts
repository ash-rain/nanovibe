import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useSetupStore } from '@/stores/setup'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: async () => {
      const setupStore = useSetupStore()
      if (!setupStore.loaded) {
        await setupStore.fetchState()
      }
      if (setupStore.currentStep === 'complete') {
        return '/app/dashboard'
      }
      return '/setup/welcome'
    },
  },

  // Setup routes
  {
    path: '/setup',
    component: () => import('@/views/setup/WizardLayout.vue'),
    children: [
      {
        path: '',
        redirect: '/setup/welcome',
      },
      {
        path: 'welcome',
        name: 'setup-welcome',
        component: () => import('@/views/setup/StepWelcome.vue'),
      },
      {
        path: 'system-check',
        name: 'setup-system-check',
        component: () => import('@/views/setup/StepSystemCheck.vue'),
      },
      {
        path: 'cloudflare',
        name: 'setup-cloudflare',
        component: () => import('@/views/setup/StepCloudflare.vue'),
      },
      {
        path: 'github',
        name: 'setup-github',
        component: () => import('@/views/setup/StepGitHub.vue'),
      },
      {
        path: 'providers',
        name: 'setup-providers',
        component: () => import('@/views/setup/StepProviders.vue'),
      },
      {
        path: 'opencode',
        name: 'setup-opencode',
        component: () => import('@/views/setup/StepOpenCode.vue'),
      },
      {
        path: 'nanoclaw',
        name: 'setup-nanoclaw',
        component: () => import('@/views/setup/StepNanoClaw.vue'),
      },
      {
        path: 'complete',
        name: 'setup-complete',
        component: () => import('@/views/setup/StepComplete.vue'),
      },
    ],
    beforeEnter: async (to, _from, next) => {
      const setupStore = useSetupStore()
      if (!setupStore.loaded) {
        await setupStore.fetchState()
      }
      if (setupStore.currentStep === 'complete' && to.path !== '/setup/complete') {
        next('/app/dashboard')
      } else {
        next()
      }
    },
  },

  // App routes
  {
    path: '/app',
    component: () => import('@/components/layout/AppShell.vue'),
    beforeEnter: async (_to, _from, next) => {
      const setupStore = useSetupStore()
      if (!setupStore.loaded) {
        await setupStore.fetchState()
      }
      if (setupStore.currentStep !== 'complete') {
        next('/setup/welcome')
      } else {
        next()
      }
    },
    children: [
      {
        path: '',
        redirect: '/app/dashboard',
      },
      {
        path: 'dashboard',
        name: 'dashboard',
        component: () => import('@/views/DashboardView.vue'),
      },
      {
        path: 'ide/:projectId?',
        name: 'ide',
        component: () => import('@/views/IDEView.vue'),
      },
      {
        path: 'agent',
        name: 'agent',
        component: () => import('@/views/AgentView.vue'),
      },
      {
        path: 'projects',
        name: 'projects',
        component: () => import('@/views/ProjectsView.vue'),
      },
      {
        path: 'github',
        name: 'github',
        component: () => import('@/views/GitHubView.vue'),
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/views/SettingsView.vue'),
      },
    ],
  },

  // GitHub OAuth callback (redirected server-side, but handle here if needed)
  {
    path: '/auth/github/callback',
    redirect: '/app/dashboard',
  },

  // SPA catch-all
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) return savedPosition
    return { top: 0 }
  },
})

export default router
