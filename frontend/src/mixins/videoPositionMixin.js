import {getStoredPresenter, getStoredVideoPosition, VIDEO_POSITION_AUTO, VIDEO_POSITION_TOP} from "@/store/localStore";
import {videochat_name} from "@/router/routes";

export default () => {
    return {
        methods: {
            videoIsOnTopPlain(value) {
                if (value == VIDEO_POSITION_AUTO) {
                    return true // both mobile and desktop
                } else {
                    return value == VIDEO_POSITION_TOP;
                }
            },
            videoIsOnTop() {
              const stored = this.chatStore.videoPosition;
              return this.videoIsOnTopPlain(stored);
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
              return true;
            },
            initPositionAndPresenter() {
                this.chatStore.videoPosition = getStoredVideoPosition();
                this.chatStore.presenterEnabled = getStoredPresenter();
            },
        }
    }
}
