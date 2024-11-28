<template>
  <div class="video-buttons-control">
    <v-btn variant="plain" icon v-if="audioPublication != null" @click="doMuteAudio(!audioMute)" :title="audioMute ? $vuetify.locale.t('$vuetify.unmute_audio') : $vuetify.locale.t('$vuetify.mute_audio')"><v-icon size="x-large" :class="['video-container-element-control-item', chatStore.muteAudioBlink && audioMute ? 'info-blink' : '']">{{ audioMute ? 'mdi-microphone-off' : 'mdi-microphone' }}</v-icon></v-btn>
    <v-btn variant="plain" icon v-if="videoPublication != null" @click="doMuteVideo(!videoMute)" :title="videoMute ? $vuetify.locale.t('$vuetify.unmute_video') : $vuetify.locale.t('$vuetify.mute_video')"><v-icon size="x-large" class="video-container-element-control-item">{{ videoMute ? 'mdi-video-off' : 'mdi-video' }} </v-icon></v-btn>
    <v-btn variant="plain" icon @click="onEnterFullscreen" :title="$vuetify.locale.t('$vuetify.fullscreen')"><v-icon size="x-large" class="video-container-element-control-item">mdi-arrow-expand-all</v-icon></v-btn>
  </div>
</template>

<script>
export default {
  data() {
    return {
      audioPublication: null, // TODO set
      videoPublication: null,
      audioMute: true,
      videoMute: true,
    }
  },
  methods: {
    setAudioStream(micPub, micEnabled) {
      this.setDisplayAudioMute(!micEnabled);
      this.audioPublication = micPub;
    },
    setVideoStream(cameraPub, cameraEnabled) {
      this.setVideoMute(!cameraEnabled);
      this.videoPublication = cameraPub;
    },
    setVideoMute(newState) {
      this.videoMute = newState;
    },
    setDisplayAudioMute(b) {
      this.audioMute = b;
    },

    doMuteAudio(requestedState) {
      // TODO emit
      this.setDisplayAudioMute(requestedState);
      this.chatStore.muteAudioBlink = false;
    },
    doMuteVideo(requestedState) {
      // TODO emit
      this.setVideoMute(requestedState);
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