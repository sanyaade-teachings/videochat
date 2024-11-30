<template>
  <div :class="videoButtonsControlClass">
    <v-btn variant="plain" icon v-if="chatStore.canShowMicrophoneButton" @click="doMuteAudio(!chatStore.localMicrophone)" :title="!chatStore.localMicrophone ? $vuetify.locale.t('$vuetify.unmute_audio') : $vuetify.locale.t('$vuetify.mute_audio')"><v-icon size="x-large" class="video-container-element-control-item">{{ !chatStore.localMicrophone ? 'mdi-microphone-off' : 'mdi-microphone' }}</v-icon></v-btn>
    <v-btn variant="plain" icon v-if="chatStore.canShowVideoButton" @click="doMuteVideo(!chatStore.localVideo)" :title="!chatStore.localVideo ? $vuetify.locale.t('$vuetify.unmute_video') : $vuetify.locale.t('$vuetify.mute_video')"><v-icon size="x-large" class="video-container-element-control-item">{{ !chatStore.localVideo ? 'mdi-video-off' : 'mdi-video' }} </v-icon></v-btn>
    <v-btn variant="plain" icon @click="onEnterFullscreen" :title="$vuetify.locale.t('$vuetify.fullscreen')"><v-icon size="x-large" class="video-container-element-control-item">mdi-arrow-expand-all</v-icon></v-btn>
  </div>
</template>

<script>
import {mapStores} from "pinia";
import {useChatStore} from "@/store/chatStore.js";
import videoPositionMixin from "@/mixins/videoPositionMixin.js";

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
        return "video-buttons-control-horizontal"
      } else if (this.videoIsVertical())  {
        if (!this.chatStore.presenterEnabled) {
          return "video-buttons-control-vertical"
        } else {
          return "video-buttons-control-horizontal"
        }
      } else {
        return null;
      }
    }
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
  }
}
</script>


<style scoped lang="stylus">

.video-buttons-control-horizontal {
  position: absolute;
  background: #f00;
  bottom 10px
  z-index 20
}

.video-buttons-control-vertical {
  margin-left: 10px;
  position: absolute;
  background: #f00;
  display: flex;
  flex-direction: column;
  z-index 20
}

</style>