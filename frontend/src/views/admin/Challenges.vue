<template>
  <section>
    <b-table :data="challenges">
      <template slot-scope="props">
        <b-table-column field="name" label="name">{{
          props.row.name
        }}</b-table-column>
        <b-table-column field="flag" label="flag">{{
          props.row.flag
        }}</b-table-column>
        <b-table-column field="solve count" label="solve count">{{
          props.row.solveteams.length
        }}</b-table-column>
        <b-table-column field="score" label="score">{{
          props.row.score
        }}</b-table-column>
        <b-table-column field="status" label="status">
          <b-switch v-model="props.row.is_open">
            <template v-if="props.row.is_open">Opened</template>
            <template v-else>Closed</template>
          </b-switch>
        </b-table-column>
      </template>
    </b-table>
    <div class="is-clearfix">
      <div class="is-pulled-right buttons">
        <b-button @click="applyStatus">Save</b-button>
      </div>
    </div>
  </section>
</template>

<script>
import API from "../../api";
import { handleError } from "../../util";

export default {
  data() {
    return {
      challenges: []
    };
  },
  methods: {
    getChallenges() {
      API.get("/admin/challenges").then(r => {
        this.challenges = r.data.challenges;
      });
    },
    applyStatus() {
      const challenges = this.challenges.map(c => {
        return {
          id: c.id,
          is_open: c.is_open
        };
      });

      API.post("/admin/set-challenges-status", {
        challenges: challenges
      })
        .then(() => {
          this.getChallenges();
        })
        .catch(e => handleError(e));
    }
  },
  mounted() {
    this.getChallenges();
  }
};
</script>
