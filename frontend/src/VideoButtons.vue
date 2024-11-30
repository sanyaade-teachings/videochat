<template>
  <div class="video-buttons-control">
    <v-btn variant="plain" icon v-if="chatStore.canShowMicrophoneButton" @click="doMuteAudio(!chatStore.localMicrophone)" :title="!chatStore.localMicrophone ? $vuetify.locale.t('$vuetify.unmute_audio') : $vuetify.locale.t('$vuetify.mute_audio')"><v-icon size="x-large" :class="['video-container-element-control-item', chatStore.muteAudioBlink && !chatStore.localMicrophone ? 'info-blink' : '']">{{ !chatStore.localMicrophone ? 'mdi-microphone-off' : 'mdi-microphone' }}</v-icon></v-btn>
    <v-btn variant="plain" icon v-if="chatStore.canShowVideoButton" @click="doMuteVideo(!chatStore.localVideo)" :title="!chatStore.localVideo ? $vuetify.locale.t('$vuetify.unmute_video') : $vuetify.locale.t('$vuetify.mute_video')"><v-icon size="x-large" class="video-container-element-control-item">{{ !chatStore.localVideo ? 'mdi-video-off' : 'mdi-video' }} </v-icon></v-btn>
    <v-btn variant="plain" icon @click="onEnterFullscreen" :title="$vuetify.locale.t('$vuetify.fullscreen')"><v-icon size="x-large" class="video-container-element-control-item">mdi-arrow-expand-all</v-icon></v-btn>
  </div>
</template>

<script>
import {mapStores} from "pinia";
import {useChatStore} from "@/store/chatStore.js";

export default {
  data() {
    return {

    }
  },
  computed: {
    ...mapStores(useChatStore),
  },
  methods: {
    doMuteAudio(requestedState) {
      this.chatStore.localMicrophone = requestedState
    },
    doMuteVideo(requestedState) {
      this.chatStore.localVideo = requestedState
    },
    onEnterFullscreen(e) {
      // TODO emit
    },
  }
}
</script>


<style scoped lang="stylus">

.video-buttons-control {
  height: 100px;
  width: 400px;
  position: absolute;
  background: #f00;
  bottom 10px
  z-index 20
}

</style>