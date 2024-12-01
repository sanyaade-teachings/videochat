<template>
    <div :class="videoContainerElementClass" ref="containerRef" @mouseenter="onMouseEnter()" @mouseleave="onMouseLeave()">
        <div class="video-container-element-control" v-show="showControls">
            <v-btn variant="plain" icon v-if="isLocal && audioPublication != null" @click="doMuteAudio(!audioMute)" :title="audioMute ? $vuetify.locale.t('$vuetify.unmute_audio') : $vuetify.locale.t('$vuetify.mute_audio')"><v-icon size="x-large" class="video-container-element-control-item">{{ audioMute ? 'mdi-microphone-off' : 'mdi-microphone' }}</v-icon></v-btn>
            <v-btn variant="plain" icon v-if="isLocal && videoPublication != null" @click="doMuteVideo(!videoMute)" :title="videoMute ? $vuetify.locale.t('$vuetify.unmute_video') : $vuetify.locale.t('$vuetify.mute_video')"><v-icon size="x-large" class="video-container-element-control-item">{{ videoMute ? 'mdi-video-off' : 'mdi-video' }} </v-icon></v-btn>
            <v-btn variant="plain" icon @click="onEnterFullscreen" :title="$vuetify.locale.t('$vuetify.fullscreen')"><v-icon size="x-large" class="video-container-element-control-item">mdi-arrow-expand-all</v-icon></v-btn>
            <v-btn variant="plain" icon v-if="isLocal" @click="onClose()" :title="$vuetify.locale.t('$vuetify.close')"><v-icon size="x-large" class="video-container-element-control-item">mdi-close</v-icon></v-btn>
            <v-btn variant="plain" icon v-if="!isLocal && canVideoKick" @click="kickRemote()" :title="$vuetify.locale.t('$vuetify.kick')"><v-icon size="x-large" class="video-container-element-control-item">mdi-block-helper</v-icon></v-btn>
            <v-btn variant="plain" icon v-if="!isLocal && canAudioMute" @click="forceMuteRemote()" :title="$vuetify.locale.t('$vuetify.force_mute')"><v-icon size="x-large" class="video-container-element-control-item">mdi-microphone-off</v-icon></v-btn>
        </div>
        <img v-show="avatarIsSet && videoMute" @click="onClick" :class="videoElementClass" :src="avatar"/>
        <video v-show="!videoMute || !avatarIsSet" @click="onClick" :class="videoElementClass" :id="id" autoPlay playsInline ref="videoRef"/>
        <p v-bind:class="[speaking ? 'video-container-element-caption-speaking' : '', 'video-container-element-caption']">{{ userName }} <v-icon v-if="audioMute">mdi-microphone-off</v-icon></p>
    </div>
</template>

<script>

import {hasLength, isFullscreen, loadingMessage} from "@/utils";
import axios from "axios";
import {mapStores} from "pinia";
import {useChatStore} from "@/store/chatStore";
import videoPositionMixin from "@/mixins/videoPositionMixin.js";
import speakingMixin from "@/mixins/speakingMixin.js";

export default {
	  name: 'UserVideo',

    mixins: [
        videoPositionMixin(),
        speakingMixin(),
    ],

    data()  {
	    return {
            userName: loadingMessage,
            audioMute: true,
            errorDescription: null,
            avatar: "",
            videoMute: true,
            userId: null,
            showControls: false,
            audioPublication: null,
            videoPublication: null,
      }
    },

    props: {
        id: {
            type: String
        },
        localVideoProperties: {
            type: Object
        },
        initialShowControls: {
            type: Boolean
        },
    },

    methods: {
        setAudioStream(micPub, micEnabled) {
            console.info("Setting source audio for videoRef=", this.$refs.videoRef, " track=", micPub, " audio tag id=", this.id, ", enabled=", micEnabled);

            this.setDisplayAudioMute(!micEnabled);
            this.audioPublication = micPub;
            if (!this.localVideoProperties) { // we don't need to hear own audio
                micPub?.audioTrack?.attach(this.$refs.videoRef);
            }
        },
        setVideoStream(cameraPub, cameraEnabled) {
            console.info("Setting source video for videoRef=", this.$refs.videoRef, " track=", cameraPub, " video tag id=", this.id, ", enabled=", cameraEnabled);
            this.setDisplayVideoMute(!cameraEnabled);
            this.videoPublication = cameraPub;

            cameraPub?.videoTrack?.attach(this.$refs.videoRef);
        },
        getVideoStream() {
            return this.videoPublication
        },
        getAudioStream() {
            return this.audioPublication
        },
        getVideoStreamId() {
            return this.videoPublication?.trackSid;
        },
        getAudioStreamId() {
            return this.audioPublication?.trackSid;
        },
        getId() {
            return this.$props.id;
        },
        getVideoElement() {
            return this?.$refs?.videoRef;
        },
        setUserName(u) {
            this.userName = u;
        },
        getUserName() {
            return this.userName;
        },
        setDisplayAudioMute(newState, skipStoreUpdate) {
            this.audioMute = newState;

            if (this.isLocal && !skipStoreUpdate) { // skipStoreUpdate prevents infinity recursion
                this.chatStore.localMicrophone = !newState
            }
        },
        onEnterFullscreen(e) {
            const elem = this.$refs.containerRef;

            if (isFullscreen()) {
                document.exitFullscreen();
            } else {
                elem.requestFullscreen();
            }
        },
        setAvatar(a) {
            this.avatar = a;
        },
        getAvatar() {
            return this.avatar;
        },
        setDisplayVideoMute(newState, skipStoreUpdate) {
            this.videoMute = newState;

            if (this.isLocal && !skipStoreUpdate) { // skipStoreUpdate prevents infinity recursion
              this.chatStore.localVideo = !newState
            }
        },
        getUserId() {
            return this.userId;
        },
        setUserId(id) {
            this.userId = id;
        },
        doMuteAudio(requestedState, skipStoreUpdate) {
            if (requestedState) {
                this.audioPublication?.mute();
            } else {
                this.audioPublication?.unmute();
            }
            this.setDisplayAudioMute(requestedState, skipStoreUpdate);
        },
        doMuteVideo(requestedState, skipStoreUpdate) {
            if (requestedState) {
                this.videoPublication?.mute();
            } else {
                this.videoPublication?.unmute();
            }
            this.setDisplayVideoMute(requestedState, skipStoreUpdate);
        },
        onClose() {
            this.localVideoProperties.localParticipant.unpublishTrack(this.videoPublication?.videoTrack);
            this.localVideoProperties.localParticipant.unpublishTrack(this.audioPublication?.audioTrack);
        },
        isComponentLocal() {
            return this.isLocal
        },
        getVideoSource() {
          return this.videoPublication?.source
        },
        onMouseEnter() {
            if (!this.isMobile()) {
                this.showControls = true;
            }
        },
        onMouseLeave() {
            if (!this.isMobile()) {
                this.showControls = false;
            }
        },
        kickRemote() {
            axios.put(`/api/video/${this.chatStore.chatDto.id}/kick?userId=${this.userId}`)
        },
        forceMuteRemote() {
            axios.put(`/api/video/${this.chatStore.chatDto.id}/mute?userId=${this.userId}`)
        },
        onClick(e) {
          if (!this.videoIsGallery()) {
            e.stopPropagation()
          }
          this.showControls =! this.showControls
        },
    },
    computed: {
        ...mapStores(useChatStore),
        avatarIsSet() {
            return hasLength(this.avatar);
        },
        isLocal() {
            return !!this.localVideoProperties;
        },
        isChangeable() {
            return this.localVideoProperties && !this.localVideoProperties.screen;
        },
        canVideoKick() { // only on remote
          return !this.isLocal && this.chatStore.canVideoKickParticipant(this.userId)
        },
        canAudioMute() { // only on remote
          return !this.isLocal && this.chatStore.canAudioMuteParticipant(this.userId)
        },
        videoContainerElementClass() {
          const ret = ['video-container-element'];
          if (this.videoIsHorizontal()) {
            ret.push('video-container-element-position-horizontal');
          } else if (this.videoIsGallery()) {
            ret.push('video-container-element-position-gallery');
          } else {
            ret.push('video-container-element-position-vertical');
          }
          return ret;
        },
        videoElementClass() {
          const ret = ['video-element'];
          if (this.videoIsHorizontal()) {
            ret.push('video-element-horizontal');
          } else if (this.videoIsGallery()) {
            ret.push('video-element-gallery');
          } else {
            ret.push('video-element-vertical');
          }
          return ret;
        },
    },
    mounted(){
        this.showControls = this.initialShowControls;
    },
};
</script>

<style lang="stylus" scoped>
    .video-container-element {
        position relative
        display flex
        flex-direction column
        align-items: center;
        //width: fit-content
        //block-size: fit-content
        //box-sizing: content-box
    }

    .video-container-element-position-horizontal {
        height 100%
    }

    .video-container-element-position-vertical {
        width 100%
    }

    .video-container-element-position-gallery {
        width: var(--width);
        height: var(--height);
        background-color: #3a3a3e;
    }

    .video-container-element:nth-child(even) {
        background #d5fdd5;
    }

    .video-container-element:nth-child(odd) {
        background #e4efff;
    }

    .video-element {
        // object-fit: contain;
        //box-sizing: border-box;
        object-fit: cover;
        z-index 2
    }

    .video-element-horizontal {
      height: 100% !important
      min-width: 100% !important
    }

    .video-element-vertical {
      height: 100% !important
      min-width: 100% !important
      width: 100% !important
    }

    .video-element-gallery {
      height: 100%;
      width: 100%;
    }

    .video-container-element-control {
        width 100%
        z-index 3
        position: absolute
    }

    .video-container-element-control-item {
        z-index 4
        text-shadow: -2px 0 white, 0 2px white, 2px 0 white, 0 -2px white;
    }

    .video-container-element-caption {
        z-index 2
        display inherit
        margin: 0;
        left 0.4em
        bottom 0.4em
        position: absolute
        background rgba(255, 255, 255, 0.65)
        max-width: calc(100% - 1em) // still needed for thin (vertical) video on mobile - it prevents bulging
        padding-left 0.3em
        padding-right 0.3em
        border-radius 4px
        //word-wrap: break-word;
        //overflow-wrap: break-all
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }

    .video-container-element-caption-speaking {
        text-shadow: -2px 0 #9cffa1, 0 2px #9cffa1, 2px 0 #9cffa1, 0 -2px #9cffa1;
    }

</style>
