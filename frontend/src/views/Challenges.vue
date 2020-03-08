<template>
  <section>
    <form
      class="column is-half is-offset-one-quarter"
      style="display: flex;"
      @submit.prevent="submitFlag"
    >
      <b-input
        v-model="flag"
        autocomplete="off"
        placeholder="flag here"
        style="flex: 1;"
      ></b-input>
      <b-button tag="input" native-type="submit" value="Submit"></b-button>
    </form>

    <section v-for="pair in challengesWithCategory" :key="pair[0]">
      <h2 class="subtitle is-2">{{ pair[0] }}</h2>
      <section
        class="challenge box"
        :class="{ solved: c.solveteams.includes(team.id) }"
        v-for="c in pair[1]"
        :key="c.id"
        @click="showInModal(c)"
      >
        <div>
          <p class="subtitle is-4">{{ c.name }}</p>
          <p class="title is-3">{{ c.score }}</p>
          <section class="tags">
            <p class="tag" v-for="t in challengeTags(c)" :key="t">{{ t }}</p>
          </section>
        </div>
      </section>
    </section>

    <div class="modal" :class="{ 'is-active': showModal }" v-if="showModal">
      <div class="modal-background" @click="showModal = false"></div>
      <div class="modalChallenge">
        <h2 class="title is-3" v-if="modalChallenge.solved">
          {{ modalChallenge.name }} (SOLVED)
        </h2>
        <h2 class="title is-3" v-else>
          {{ modalChallenge.name }}
        </h2>
        <section class="tags">
          <p class="tag" v-for="t in challengeTags(modalChallenge)" :key="t">
            {{ t }}
          </p>
        </section>
        <div class="modalDescription">
          <p
            class="column is-half is-offset-one-quarter"
            v-html="modalChallenge.description"
          ></p>
        </div>

        <p class="has-text-right">author: {{ modalChallenge.author }}</p>
        <div>
          <b-button
            tag="a"
            :href="url"
            target="_blank"
            download
            v-for="url in modalChallenge.attachments"
            :key="url"
            >Download Attachments</b-button
          >
        </div>
      </div>
      <button
        class="modal-close is-large"
        aria-label="close"
        @click="showModal = false"
      ></button>
    </div>
  </section>
</template>

<script>
import Vue from "vue";
import API from "../api";
import { handleError, showMessage } from "../util";
import lodash from "lodash";

export default {
  data() {
    return {
      challenges: {},
      team: null,
      showModal: false,
      modalChallenge: null,
      flag: ""
    };
  },
  mounted() {
    API.get("/team")
      .then(r => {
        this.team = r.data.team;
      })
      .catch(e => handleError(this, e));
    API.get("/challenges")
      .then(r => {
        this.challenges = r.data.challenges.reduce((map, c) => {
          map[c.id] = c;
          return map;
        }, {});
      })
      .catch(e => handleError(this, e));

    this.$eventHub.$on("challengeUpdate", c => {
      Vue.set(this.challenges, c.id, c);
      this.$forceUpdate();
    });
    this.$eventHub.$on("challengeClose", cid => {
      Vue.delete(this.challenges, cid);
      this.$forceUpdate();
    });
  },
  methods: {
    challengeTags(c) {
      return [].concat(c.difficulty, c.tags || []);
    },
    showInModal(c) {
      this.modalChallenge = c;
      this.showModal = true;
    },
    submitFlag() {
      API.post("/submit", {
        flag: this.flag
      })
        .then(r => showMessage(this, r.data.message))
        .catch(e => handleError(this, e));
    }
  },
  computed: {
    challengesWithCategory() {
      return lodash(this.challenges)
        .groupBy("category")
        .toPairs()
        .orderBy()
        .value();
    }
  }
};
</script>

<style lang="scss" scoped>
@import "../style.scss";

.flagSubmitInput {
  flex: 1;
}

.challenge {
  display: inline-flex;
  padding: 0.75rem;
  margin: 0.25rem 0.75rem;
  border-radius: 0;
  cursor: pointer;
}
.solved {
  filter: opacity(0.25);
}

.tags {
  font-size: 1rem;

  .tag {
    display: inline-flex;
    margin-right: 0.75rem;
    background-color: rgba(0, 0, 0, 0.2);
  }
}
.modalChallenge {
  display: flex;
  flex-direction: column;

  position: fixed;
  background-color: $body-background-color;
  padding: 0.75rem;

  left: 25%;
  right: 25%;
}
.modalDescription {
  background-color: rgba(0, 0, 0, 0.2);
  flex: 1;
}
</style>
