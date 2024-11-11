import {getStoredPresenter, getStoredVideoPosition, VIDEO_POSITION_AUTO, VIDEO_POSITION_HORIZONTAL} from "@/store/localStore";
import {videochat_name} from "@/router/routes";

export default () => {
    return {
        methods: {
            videoIsHorizontalPlain(value) {
                if (value == VIDEO_POSITION_AUTO) {
                    return true // both mobile and desktop
                } else {
                    return value == VIDEO_POSITION_HORIZONTAL;
                }
            },
            videoIsHorizontal() {
              const stored = this.chatStore.videoPosition;
              return this.videoIsHorizontalPlain(stored);
            },
            videoIsVertical() {
              return !this.videoIsHorizontal();
            },

            isVideoRoute() {
              return this.$route.name == videochat_name
            },

            shouldShowChatList() {
              if (this.isMobile()) {
                return false;
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
