<template>
  <v-container fluid>
    <v-card 
      outlined
      :height="dimension.windowHeight-100"
    >
      <v-card-title>Job Monitor</v-card-title>

      <k-browser-2 ref="grid"
        list-mode="grid"
        list-meta="/job/gridconfig"
        list-source="/job/gets"
        :list-custom-fields="['_id','Status']"
        :list-source-parm="{itemsPerPage:10}"
        :list-show-footer="true"
        :show-new="false"
        :show-delete="false"
        :show-edit="false"
        :prevent-dbl-click="true"
        @dblClick="showItemOnDrawer"
      >
        <template v-slot:list_item__id="item">
          {{ item._id.substr(item._id.length-6)  }}
        </template>

        <template v-slot:list_item_Status="item">
          <v-chip color="primary" v-if="item.Status=='Running'" small>Running</v-chip>
          <v-chip color="green" v-if="item.Status=='Done'" small>Done</v-chip>
          <v-chip color="error" v-if="item.Status=='Stopped'" small>Stopped</v-chip>
          <v-chip color="secondary" v-if="item.Status=='New'" style="cursor:pointer" small @click="startJob(item)">Click to start</v-chip>
        </template>

        <template v-slot:list_extra_buttons="item">
          <v-btn icon color="primary" x-small @click="showDrawer(item)">
            <v-icon>mdi-playlist-check</v-icon>
          </v-btn>
        </template>
      </k-browser-2>

      <v-navigation-drawer
        v-if="drawer"
        absolute
        permanent
        right
        :width="320"
        style="z-index:10000"
      >
        <v-toolbar height="20" flat dense class="ma-2">
          <v-spacer />
          <v-icon small color="primary" class="ml-2" @click="drawer=false">mdi-arrow-expand-right</v-icon>
        </v-toolbar>
        
        <div class="ml-5 mr-5 mt-5">
          <h1 style="font-size:1.2em">{{ item._id }}</h1>
          <h2 style="font-size:1em">
            {{ item.PipeID }}
            <v-chip x-small color="secondary" v-if="item.Status=='Running'">{{ item.Status }}</v-chip> 
            <v-chip x-small color="error" v-if="item.Status=='Stopped'">{{ item.Status }}</v-chip> 
            <v-chip x-small color="green" v-if="item.Status=='Done'">{{ item.Status }}</v-chip> 
            {{ item.InitialData.length }} data, 
            {{ $moment(item.Created).format('DD-MM-YYYY hh:mm:ss az') }}</h2>

          <div class="mt-2">
            <v-btn class="secondary mr-1" small v-if="item.Status!='Running'" @click="startJob(item)">Run Job</v-btn>
            <v-btn class="error mr-1" small v-if="item.Status=='Running'" @click="stopJob(item)">Stop</v-btn>
          </div>

          <div class="mt-2">
            <v-list>
              <v-list-item v-for="wid in Object.keys(item.Workers)" :key="'worker'+wid">
                <v-list-item-icon>
                  <v-icon color="warning" small>mdi-circle</v-icon>
                </v-list-item-icon>  
                <v-list-item-content>
                  <v-list-item-title>{{ wid }}</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </div>
        </div>
      </v-navigation-drawer>
    </v-card>
  </v-container>
</template>

<script>
import KBrowser2 from '@shared/components/k-vue/KBrowser2.vue'
import dimension from '@/mixin/dimension.js'
export default {
  components: { KBrowser2 },
  name: 'DFJob',
  mixins: [dimension],
  data () {
    return {
      drawer: false,
      item: {}
    }
  },

  mounted () {
    this.refreshGrid()
  },

  methods: {
    refreshGrid () {
      this.$refs.grid.refresh()
      window.setTimeout(this.refreshGrid, 5000)
    },

    showItemOnDrawer (item) {
      this.item = item
      this.drawer = true
    },

    showDrawer (item) {
      this.item = item
      this.drawer = !this.drawer
    },

    startJob (item) {
      this.$refs.grid.refresh()

      this.$axios.post('/coordinator/startjob',item._id).then(r => {
        // reload grid
        this.$refs.grid.refresh()

        // get item
        this.$axios.post("/job/get",[item._id]).then(r=>{
          this.item = r.data
        })
      },
      e => this.$tool.error(e))
    },

    stopJob (item) {
      this.$refs.grid.refresh()

      this.$axios.post('/coordinator/stopjob',{
        JobID: item._id, 
        Status: 'Stopped'
      }).then(r => {
        // reload grid
        this.$refs.grid.refresh()

        // get item
        this.$axios.post("/job/get",[item._id]).then(r=>{
          this.item = r.data
        })
      },
      e => this.$tool.error(e))
    }
  }
}
</script>