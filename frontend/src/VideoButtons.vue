<template>
  <div :class="videoButtonsControlClass">
    <template v-if="chatStore.showCallManagement && chatStore.isInCall()">
      <v-btn variant="plain" icon :loading="chatStore.initializingVideoCall" @click.stop.prevent="stopCall()" :title="$vuetify.locale.t('$vuetify.leave_call')">
        <v-icon size="x-large" :class="chatStore.shouldPhoneBlink ? 'call-blink' : 'text-red'">mdi-phone</v-icon>
      </v-btn>
    </template>
    <v-btn variant="plain" icon v-if="chatStore.canShowMicrophoneButton" @click.stop.prevent="doMuteAudio(!chatStore.localMicrophone)" :title="!chatStore.localMicrophone ? $vuetify.locale.t('$vuetify.unmute_audio') : $vuetify.locale.t('$vuetify.mute_audio')"><v-icon size="x-large">{{ !chatStore.localMicrophone ? 'mdi-microphone-off' : 'mdi-microphone' }}</v-icon></v-btn>
    <v-btn variant="plain" icon v-if="chatStore.canShowVideoButton" @click.stop.prevent="doMuteVideo(!chatStore.localVideo)" :title="!chatStore.localVideo ? $vuetify.locale.t('$vuetify.unmute_video') : $vuetify.locale.t('$vuetify.mute_video')"><v-icon size="x-large">{{ !chatStore.localVideo ? 'mdi-video-off' : 'mdi-video' }} </v-icon></v-btn>
    <template v-if="!isMobile()">
      <v-btn variant="plain" icon @click.stop.prevent="addScreenSource()" :title="$vuetify.locale.t('$vuetify.screen_share')">
        <v-icon size="x-large">mdi-monitor-screenshot</v-icon>
      </v-btn>
    </template>
    <v-btn variant="plain" icon @click.stop.prevent="onEnterFullscreen" :title="$vuetify.locale.t('$vuetify.fullscreen')"><v-icon size="x-large">mdi-arrow-expand-all</v-icon></v-btn>

    <v-select
        class="video-position-select"
        :items="positionItems"
        density="compact"
        hide-details
        @update:modelValue="changeVideoPosition"
        v-model="chatStore.videoPosition"
        variant="plain"
    ></v-select>

  </div>
</template>

<script>
import {mapStores} from "pinia";
import {useChatStore} from "@/store/chatStore.js";
import videoPositionMixin from "@/mixins/videoPositionMixin.js";
import {stopCall} from "@/utils.js";
import bus, {ADD_SCREEN_SOURCE} from "@/bus/bus.js";
import {
  setStoredVideoPosition,
  VIDEO_POSITION_AUTO, VIDEO_POSITION_GALLERY,
  VIDEO_POSITION_HORIZONTAL,
  VIDEO_POSITION_VERTICAL
} from "@/store/localStore.js";

export default {
  mixins: [
    videoPositionMixin(),
  ],
  data() {
    return {

    }
  },
  computed: {
    ...mapStores(useChatStore),
    videoButtonsControlClass() {
      if (this.videoIsHorizontal() || this.videoIsGallery()) {
        return ["video-buttons-control", "video-buttons-control-horizontal"]
      } else if (this.videoIsVertical())  {
        if (!this.chatStore.presenterEnabled) {
          const vbcv = this.isMobile() ? "video-buttons-control-vertical-mobile" : "video-buttons-control-vertical";
          return ["video-buttons-control", vbcv]
        } else {
          return ["video-buttons-control", "video-buttons-control-horizontal"]
        }
      } else {
        return null;
      }
    },
    positionItems() {
      return [VIDEO_POSITION_AUTO, VIDEO_POSITION_HORIZONTAL, VIDEO_POSITION_VERTICAL, VIDEO_POSITION_GALLERY]
    },
  },
  methods: {
    doMuteAudio(requestedState) {
      this.chatStore.localMicrophone = requestedState
    },
    doMuteVideo(requestedState) {
      this.chatStore.localVideo = requestedState
    },
    onEnterFullscreen(e) {
      this.$emit("requestFullScreen");
    },

    stopCall() {
      stopCall(this.chatStore, this.$route, this.$router);
    },
    addScreenSource() {
      bus.emit(ADD_SCREEN_SOURCE);
    },
    changeVideoPosition(v) {
      this.chatStore.videoPosition = v;
      setStoredVideoPosition(v);
    },
  }
}
</script>


<style scoped lang="stylus">

.video-buttons-control {
  background rgba(255, 255, 255, 0.65)
  padding-left 0.3em
  padding-right 0.3em
  border-radius 4px
  display: flex;
}

.video-buttons-control-horizontal {
  position: absolute;
  bottom 16px
  z-index 20
}

.video-buttons-control-vertical {
  margin-left: 10px;
  position: absolute;
  display: flex;
  flex-direction: column;
  z-index 20
}

.video-buttons-control-vertical-mobile {
  position: absolute;
  display: flex;
  flex-direction: column;
  z-index 20
  left: 10px;
}

.video-position-select {
  margin-top auto
  margin-bottom auto
  display: inline-flex
  align-self: center
}

</style>

<style lang="stylus">
.video-position-select {
  .v-field__input {
    min-height: unset;
    min-width: unset;
    padding 0 !important
    margin 0 !important
    position: relative;
  }

  div.v-field__append-inner {
      padding 0 !important
  }
}
</style>