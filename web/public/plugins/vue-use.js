import Vue from 'vue'
import draggable from 'vuedraggable'
import JsonEditor from '@shared/components/vue-json-edit/JsonEditor.vue'

Vue.component('draggable',draggable)
Vue.component('JsonEditor', JsonEditor)