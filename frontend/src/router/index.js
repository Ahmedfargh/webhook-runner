import { createRouter, createWebHistory } from 'vue-router'
import AppLayout from '../components/layout/AppLayout.vue'
import DashboardView from '../views/DashboardView.vue'
import UsersView from '../views/UsersView.vue'
import AdminsView from '../views/AdminsView.vue'
import RolesView from '../views/RolesView.vue'
import PermissionsView from '../views/PermissionsView.vue'
import SystemTopologyView from '../views/SystemTopologyView.vue'
import PlansView from '../views/PlansView.vue'
import MySubscriptionView from '../views/MySubscriptionView.vue'
import InvoicesView from '../views/InvoicesView.vue'
import AdminBillingView from '../views/AdminBillingView.vue'
import AppsView from '../views/AppsView.vue'
import WebhookLogsView from '../views/WebhookLogsView.vue'
import LoginView from '../views/LoginView.vue'
import RegisterView from '../views/RegisterView.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: LoginView,
    meta: { public: true },
  },
  {
    path: '/register',
    name: 'Register',
    component: RegisterView,
    meta: { public: true },
  },
  {
    path: '/',
    component: AppLayout,
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: DashboardView,
      },
      {
        path: 'apps',
        name: 'Apps',
        component: AppsView,
      },
      {
        path: 'webhooks/logs',
        name: 'WebhookLogs',
        component: WebhookLogsView,
      },
      {
        path: 'users',
        name: 'Users',
        component: UsersView,
        meta: { requiresAdmin: true },
      },
      {
        path: 'admins',
        name: 'Admins',
        component: AdminsView,
        meta: { requiresAdmin: true },
      },
      {
        path: 'roles',
        name: 'Roles',
        component: RolesView,
        meta: { requiresAdmin: true },
      },
      {
        path: 'permissions',
        name: 'Permissions',
        component: PermissionsView,
        meta: { requiresAdmin: true },
      },
      {
        path: 'plans',
        name: 'Plans',
        component: PlansView,
      },
      {
        path: 'subscription',
        name: 'MySubscription',
        component: MySubscriptionView,
      },
      {
        path: 'invoices',
        name: 'Invoices',
        component: InvoicesView,
      },
      {
        path: 'admin/billing',
        name: 'AdminBilling',
        component: AdminBillingView,
        meta: { requiresAdmin: true },
      },
      {
        path: 'topology',
        name: 'Topology',
        component: SystemTopologyView,
        meta: { requiresAdmin: true },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation Guard: Redirect to /login if not authenticated, or block unauthorized admin access
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('auth_token')
  const userStr = localStorage.getItem('auth_user')
  let user = null
  try {
    user = userStr ? JSON.parse(userStr) : null
  } catch (e) {
    user = null
  }
  const isAdmin = user?.role === 'admin' || user?.role === 'administrator'

  if (!to.meta.public && !token) {
    next('/login')
  } else if (to.meta.public && token) {
    next('/')
  } else if (to.meta.requiresAdmin && !isAdmin) {
    next('/plans')
  } else if (to.path === '/subscription' && isAdmin) {
    next('/admin/billing')
  } else {
    next()
  }
})

export default router
