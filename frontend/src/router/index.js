import Vue from "vue";
import VueRouter from "vue-router";
import Index from "../views/Index.vue";
import Login from "../views/Login.vue";
import Register from "../views/Register.vue";
import Challenges from "../views/Challenges.vue";
import Team from "../views/Team.vue";
import Ranking from "../views/Ranking.vue";
import PasswordResetRequest from "../views/PasswordResetRequest.vue";
import PasswordReset from "../views/PasswordReset.vue";

import Admin from "../views/Admin.vue";
import AdminConfig from "../views/admin/Config.vue";
import AdminChallenges from "../views/admin/Challenges.vue";

Vue.use(VueRouter);

const routes = [
  {
    path: "/",
    name: "Index",
    component: Index
  },
  {
    path: "/register",
    name: "Register",
    component: Register
  },
  {
    path: "/login",
    name: "Login",
    component: Login
  },
  {
    path: "/reset-request",
    name: "ResetRequest",
    component: PasswordResetRequest
  },
  {
    path: "/reset",
    name: "Reset",
    component: PasswordReset
  },
  {
    path: "/challenges",
    name: "Challenges",
    component: Challenges
  },
  {
    path: "/team/:id",
    name: "Team",
    component: Team
  },
  {
    path: "/ranking",
    name: "Ranking",
    component: Ranking
  },
  {
    path: "/admin",
    component: Admin
  },
  {
    path: "/admin/",
    component: Admin,
    children: [
      {
        path: "config",
        component: AdminConfig
      },
      {
        path: "challenges",
        component: AdminChallenges
      }
    ]
  }
];

const router = new VueRouter({
  base: process.env.BASE_URL,
  routes
});

export default router;
