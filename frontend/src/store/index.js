import Vue from "vue";
import Vuex from "vuex";

Vue.use(Vuex);

export default new Vuex.Store({
  state: {
    user: null
  },
  mutations: {
    login(state, user) {
      state.user = user;
    },
    logout(state) {
      state.user = null;
    }
  },
  actions: {
    login(context, user) {
      context.commit("login", user);
    },
    logout(context) {
      context.commit("logout");
    }
  },
  modules: {}
});
