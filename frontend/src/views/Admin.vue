<template>
  <section>
    <h1 class="is-size-4">Admin Page</h1>
    <div class="buttons">
      <b-button tag="router-link" to="/admin/config">Config</b-button>
      <b-button tag="router-link" to="/admin/challenges">Challenges</b-button>
      <button @click="updateScore">Score Refresh</button>
    </div>

    <router-view></router-view>
  </section>
</template>

<script>
import API from "../api";
import { handleError } from "../util";
export default {
  methods: {
    updateScore() {
      API.post("/admin/scoreupdate").catch(e => handleError(this, e));
    }
  },
  mounted() {
    API.get("/admin").catch(e => {
      handleError(this, e);
      this.$router.push("/");
    });
  }
};
</script>
