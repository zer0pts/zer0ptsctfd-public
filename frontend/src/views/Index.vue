<template>
  <section>
    <h1 class="is-size-1">{{ ctfName }}</h1>
    <section class="column is-offset-1">
      <p class="is-size-4">{{ startTime }} 〜 {{ endTime }}</p>
      <p class="is-size-4" v-if="willHold">CTF will start in {{ countDown }}</p>
      <p class="is-size-4" v-else-if="nowRunning">CTF now running!</p>
      <p class="is-size-4" v-else-if="hasEnd">
        CTF is over. Thanks for playing!
      </p>
    </section>
    <section class="column is-offset-1">
      <h2 class="is-size-3">[ About ]</h2>
      <p class="column">
        Welcome to zer0pts CTF 2020!<br />
        zer0pts CTF is a Jeopardy-style Capture The Flag competition hosted by
        zer0pts, a Japanese CTF team. There will be challenges of 5 categories
        (pwn, web, rev, crypto, forensics) with various difficulties. The flag
        format is <code>zer0pts{[a-zA-Z0-9_\+\!\?]+}</code> unless otherwise
        specified.<br />
        The CTF will start at March 7th 00:00 UTC and end at March 9th 00:00
        UTC.
      </p>
    </section>
    <section class="column is-offset-1">
      <h2 class="is-size-3">[ Contact ]</h2>
      <p>
        Discord:
      </p>
    </section>
    <section class="column is-offset-1">
      <h2 class="is-size-3">[ Rules ]</h2>
      <ul style="list-style-type:disc;">
        <li>
          There is no restriction on the team size, your age and nationality.
        </li>
        <li>
          The team who earns more points will be placed higher. If two teams
          have the same amount of points, the team who reached the score earlier
          will win (except for the survey*).
        </li>
        <li>
          Sharing solutions or hints with other teams during the competition is
          forbidden.
        </li>
        <li>
          Attacking the score server is forbidden. We may disqualify and ban the
          team which attacks the score server. Attacking other teams is
          forbidden as well.
        </li>
        <li>
          You are not allowed to brute-force the flag. The form will be locked
          for a while if you submit wrong flags 5 times successively.
        </li>
        <li>You may not play the CTF in multiple teams.</li>
        <li>
          You may not have multiple accounts. In case you can't log in to your
          account, please contact us in Discord.
        </li>
      </ul>
      <small
        >*In the late of the competition we will open a survey as a challenge,
        which has points as well as other challenges. However, even if the
        lower-ranked team among those who have the same amount of points solves
        the survey earlier, the originally higher-ranked teams will be placed
        higher as long as they solve the survey during the competition. Be
        noticed this is an exception to give the participants enough time to
        answer the survey while encouraging them to submit it.</small
      >
    </section>
    <section class="column is-offset-1">
      <h2 class="is-size-3">[ Prizes ]</h2>
      <ul style="list-style-type:disc;">
        <li>1st: $500</li>
        <li>2nd: $300</li>
        <li>3rd: $100</li>
      </ul>
    </section>
    <section class="column is-offset-1">
      <h2 class="is-size-3">[ Sponsors ]</h2>
      <div class="columns">
      </div>
    </section>
  </section>
</template>

<script>
import API from "../api";
import { handleError } from "../util";
import dayjs from "dayjs";

export default {
  data() {
    return {
      ctfName: "",
      startAt: null,
      endAt: null,
      now: 0
    };
  },
  methods: {
    dateFormat(ts) {
      return dayjs(ts * 1000).format("YYYY-MM-DD HH:mm:ss Z");
    }
  },
  mounted() {
    setInterval(() => {
      this.now = Math.floor(new Date().valueOf() / 1000);
    }, 1000);

    API.get("/ctf")
      .then(r => {
        if (r.data.config) {
          ({
            ctf_name: this.ctfName,
            start_at: this.startAt,
            end_at: this.endAt
          } = r.data.config);
        }
      })
      .catch(e => handleError(this, e));
  },
  computed: {
    startTime() {
      return this.dateFormat(this.startAt);
    },
    endTime() {
      return this.dateFormat(this.endAt);
    },
    willHold() {
      return this.now < this.startAt;
    },
    nowRunning() {
      return this.startAt <= this.now && this.now <= this.endAt;
    },
    hasEnd() {
      return this.now > this.endAt;
    },
    countDown() {
      const d = this.startAt - this.now;
      const days = ("" + Math.floor(d / (60 * 60 * 24))).padStart(2, "0");
      const hours = (
        "" + Math.floor((d % (60 * 60 * 24)) / (60 * 60))
      ).padStart(2, "0");
      const minutes = ("" + Math.floor((d % (60 * 60)) / 60)).padStart(2, 0);
      const seconds = ("" + Math.floor(d % 60)).padStart(2, 0);
      return days + "d " + hours + ":" + minutes + ":" + seconds;
    }
  }
};
</script>

<style scoped>
.hematite {
  height: 5em;
}
.activedefense {
  height: 7em;
}
</style>
