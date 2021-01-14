<template>
  <v-container fluid>
    <v-card 
      outlined
      :height="dimension.windowHeight-100"
    >
      <v-card-title>
        Job Log
        <v-spacer/>
        <div style="text-align:right;font-weight:none;font-size:10px">
        <b>Last Check :</b> {{ $moment(lastCheck).format('DD-MMM-yy hh:mm:ss a') }}
        </div>
      </v-card-title>

      <k-browser-2 ref="grid"
        list-mode="grid"
        list-meta="/pipelog/gridconfig"
        list-source="/pipelog/gets"
        :list-source-parm="{itemsPerPage:10}"
        :list-show-footer="true"
        :show-new="false"
        :show-delete="false"
        :show-edit="false"
        :prevent-dbl-click="true"
      >
      </k-browser-2>
    </v-card>
  </v-container>
</template>

<script>
import KBrowser2 from '@shared/components/k-vue/KBrowser2.vue'
import dimension from '@/mixin/dimension.js'
export default {
  components: { KBrowser2 },
  name: 'DFPipeLog',
  mixins: [dimension],
  data () {
    return {
      drawer: false,
      lastCheck: '',
      item: {}
    }
  },

  mounted () {
    this.refreshGrid()
  },

  methods: {
    refreshGrid () {
      this.$refs.grid.refresh()
      this.lastCheck = new Date()
      window.setTimeout(this.refreshGrid, 5000)
    }
  }
}
</script>