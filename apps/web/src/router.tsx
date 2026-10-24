import {
  createRouter,
  createRoute,
  createRootRoute,
  Outlet,
} from '@tanstack/react-router'
import { Layout } from '@/components/layout/Layout'
import { HomePage } from '@/pages/Home'
import { BlogListingPage } from '@/pages/BlogListing'
import { BlogDetailPage } from '@/pages/BlogDetail'
import { LoginPage } from '@/pages/LoginPage'
import { NotFoundPage } from '@/pages/NotFound'

const rootRoute = createRootRoute({
  component: () => (
    <Layout>
      <Outlet />
    </Layout>
  ),
  notFoundComponent: NotFoundPage,
})

const aboutRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/home',
  component: HomePage,
})

const homeRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/home',
  component: HomePage,
})

const blogListRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: BlogListingPage,
})
const blogDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/blog/$slug',
  component: BlogDetailPage,
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
  loginRoute,
  aboutRoute
])

export const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
