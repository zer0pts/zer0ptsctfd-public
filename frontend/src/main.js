import Vue from "vue";
import App from "./App.vue";
import router from "./router";
import store from "./store";
import Buefy from "buefy";
import "buefy/dist/buefy.css";
import "./style.scss";
import EventHub from "./eventHub";
Vue.config.productionTip = false;
Vue.use(Buefy);
Vue.use(EventHub);

new Vue({
  router,
  store,
  render: function(h) {
    return h(App);
  }
}).$mount("#app");
