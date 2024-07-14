import { Node, mergeAttributes } from '@tiptap/core';

// https://www.codemzy.com/blog/tiptap-video-embed-extension
const Video = Node.create({
    name: 'video', // unique name for the Node
    group: 'inline',
    selectable: true, // so we can select the video
    draggable: true, // so we can drag the video
    atom: true, // is a single unit
    inline: true,

    parseHTML() {
        return [
            {
                tag: 'video',
            },
        ]
    },
    addAttributes() {
        return {
            "src": {
                default: null
            },
            "poster": {
                default: null
            },
        }
    },
    renderHTML({ HTMLAttributes }) {
        return ['video', mergeAttributes({"class": "video-custom-class", "controls": true}, HTMLAttributes)];
    },
    addCommands() {
        return {
            setVideo: options => ({ commands }) => {
                return commands.insertContent({
                    type: this.name,
                    attrs: options,
                })
            },
        }
    },

    // https://www.codemzy.com/blog/tiptap-video-embed-extension
    addNodeView() {
        return ({ editor, node }) => {
            const div = document.createElement('div');
            div.classList.add("my-class");
            const video = document.createElement('video');
            video.src = node.attrs.src;
            div.append(video);
            return {
                dom: div,
            }
        }
    },
});

export default Video;
