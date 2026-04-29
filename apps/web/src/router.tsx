import {
  createRouter,
  createRoute,
  createRootRoute,
  Outlet,
  redirect,
} from '@tanstack/react-router'
import { Layout } from '@/components/layout/Layout'
import { HomePage } from '@/pages/Home'
import { BlogListingPage } from '@/pages/BlogListing'
import { BlogDetailPage } from '@/pages/BlogDetail'
import { EditorPage } from '@/pages/EditorPage'
import { LoginPage } from '@/pages/LoginPage'
import { NotFoundPage } from '@/pages/NotFound'
import { getToken } from '@/lib/api'

const rootRoute = createRootRoute({
  component: () => (
    <Layout>
      <Outlet />
    </Layout>
  ),
  notFoundComponent: NotFoundPage,
})

const homeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: HomePage,
})

const blogListRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/blog',
  component: BlogListingPage,
})

const blogDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/blog/$slug',
  component: BlogDetailPage,
})

const editorRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin/editor',
  component: EditorPage,
  beforeLoad: () => {
    if (!getToken()) {
      throw redirect({ to: '/admin/login' })
    }
  },
})

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/admin/login',
  component: LoginPage,
})

const routeTree = rootRoute.addChildren([
  homeRoute,
  blogListRoute,
  blogDetailRoute,
  editorRoute,
  loginRoute,
])

export const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
