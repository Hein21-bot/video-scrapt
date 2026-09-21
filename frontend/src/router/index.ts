import { createRouter, createWebHistory } from 'vue-router'
import HomePage        from '../pages/HomePage.vue'
import WatchPage       from '../pages/WatchPage.vue'
import SharePage       from '../pages/SharePage.vue'
import ActorsPage      from '../pages/ActorsPage.vue'
import ActorVideosPage from '../pages/ActorVideosPage.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/',            component: HomePage },
    { path: '/actors',      component: ActorsPage },
    { path: '/actor/:slug', component: ActorVideosPage },
    { path: '/watch/:id',   component: WatchPage },
    { path: '/s/:code',     component: SharePage },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

export default router
