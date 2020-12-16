<template>
  <v-container fluid>
    <v-card 
      outlined
      :height="windowHeight-100"
    >
      <v-card-title>Scanner</v-card-title>

      <k-browser-2
        list-mode="grid"
        list-source="/coordinator/scanners"
        list-meta="/scanner/gridconfig"
        list-save-url="/scanner/save"
        list-delete-url="/scanner/delete"
        :list-inline-editor="true"
        :list-source-parm="{itemsPerPage:-1}"
        :list-show-footer="false"
        :show-new="true"
        :show-delete="true"
      >
        <template v-slot:list_extra_buttons="item">
          <v-btn @click="showBeat(item)" icon color="primary" x-small>
            <v-icon>mdi-playlist-check</v-icon>
          </v-btn>
        </template>
      </k-browser-2>

      <v-navigation-drawer
        v-if="beatFlag"
        absolute
        permanent
        right
        :width="600"
        style="z-index:10000"
      >
        <v-toolbar height="20" flat dense class="ma-2">
          <v-toolbar-title style="font-size:1.2em">{{ item._id }}</v-toolbar-title>
          <v-spacer />
          <v-icon small color="primary" class="ml-2">mdi-playlist-check</v-icon>
          <v-icon small color="primary" class="ml-2">mdi-view-list</v-icon>
          <v-icon small color="primary" class="ml-2" @click="beatFlag=false">mdi-arrow-expand-right</v-icon>
        </v-toolbar>
        
        <div class="ml-5 mr-5 mt-5" style="height:500px">
          <v-list>
            <v-list-item v-for="(node) in item.Nodes" :key="'node-'+node._id">
              <v-list-item-action>
                <v-icon small color="green" v-if="node.Status!='Error'">mdi-circle</v-icon>
                <v-icon small color="error" v-if="node.Status=='Error'">mdi-circle</v-icon>
              </v-list-item-action>

              <v-list-item-content>
                <v-list-item-title>Node {{ node.ID }}</v-list-item-title>
                <v-list-item-subtitle>{{ $moment(node.LastUpdate).format('DD-MMM-yyyy hh:mm:ss a z') }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-list>
        </div>
      </v-navigation-drawer>
    </v-card>
  </v-container>
</template>

<script>
import KBrowser2 from '@shared/components/k-vue/KBrowser2.vue'
export default {
  components: { KBrowser2 },
  name: 'DFScanner',
  data () {
    return {
      beatFlag: false,
      windowHeight: 500,
      item: {}
    }
  },

  mounted () {
    this.$nextTick(() => {
      this.windowHeight = window.innerHeight
      window.addEventListener('resize', this.onResize)
    })
  },

  beforeDestroy() { 
    window.removeEventListener('resize', this.onResize); 
  },

  methods: {
    onResize () {
      this.windowHeight = window.innerHeight
    },
    
    showBeat (item) {
      this.item = item
      this.beatFlag = true
    }
  }
}
</script>