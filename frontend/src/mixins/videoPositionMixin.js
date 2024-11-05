import {getStoredPresenter, getStoredVideoPosition, VIDEO_POSITION_AUTO, VIDEO_POSITION_TOP} from "@/store/localStore";
import {videochat_name} from "@/router/routes";

export default () => {
    return {
        methods: {
            videoIsOnTop() {
              const stored = this.chatStore.videoPosition;
              if (stored == VIDEO_POSITION_AUTO) {
                return true // both mobile and desktop
              } else {
                return stored == VIDEO_POSITION_TOP;
              }
            },

            videoIsAtSide() {
              return !this.videoIsOnTop();
            },

            isVideoRoute() {
              return this.$route.name == videochat_name
            },

            shouldShowChatList() {
              if (this.isMobile()) {
                return false;
              }
              if (this.isVideoRoute()) {
                if (this.videoIsAtSide()) {
                  return false
                }
              }
              return true;
            },
            initPositionAndPresenter() {
                this.chatStore.videoPosition = getStoredVideoPosition();
                this.chatStore.presenterEnabled = getStoredPresenter();
            },
        }
    }
}
