import { defineConfig } from "umi";

export default defineConfig({
  routes: [
    {path: "/addAccount", component: "addAccount"},
    {path: "/login", component: "login"},
    {path: "/register", component: "register"},
    {path: "/accountManage", component: "accountManage"},
    {path: "/configure/:accountEmail", component: "configure"},
  ],
  npmClient: 'pnpm',
  esbuildMinifyIIFE: true
});
