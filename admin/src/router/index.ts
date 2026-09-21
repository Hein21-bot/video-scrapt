import { createRouter, createWebHistory } from 'vue-router'
import { isAdmin } from '../stores/auth'
import AdminLogin     from '../pages/AdminLogin.vue'
import AdminLayout    from '../pages/AdminLayout.vue'
import AdminDashboard from '../pages/AdminDashboard.vue'
import AdminVideos    from '../pages/AdminVideos.vue'
import AdminScrape    from '../pages/AdminScrape.vue'
import AdminCategories from '../pages/AdminCategories.vue'
import AdminChannels   from '../pages/AdminChannels.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/admin/login', component: AdminLogin },
    {
      path: '/admin',
      component: AdminLayout,
      meta: { requiresAdmin: true },
      children: [
        { path: '',       component: AdminDashboard },
        { path: 'videos', component: AdminVideos },
        { path: 'scrape', component: AdminScrape },
        { path: 'categories', component: AdminCategories },
        { path: 'channels', component: AdminChannels },
      ],
    },
    { path: '/:catchAll(.*)', redirect: '/admin' },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  if (to.meta.requiresAdmin && !isAdmin.value) return '/admin/login'
})

export default router
