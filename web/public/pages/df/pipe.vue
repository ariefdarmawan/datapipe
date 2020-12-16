<template>
  <v-container fluid>
    <v-card 
      :height="dimension.windowHeight-100"
      outlined>
      <v-card-title>
        Data Pipe
      </v-card-title>

      <k-browser-2
        ref ="db"
        list-mode='grid'
        list-meta='/pipe/gridconfig'
        list-source='/pipe/gets'
        form-meta='/pipe/formconfig'
        form-source='/pipe/get'
        form-save='/pipe/save'
        list-delete-url="/pipe/delete"
        :list-custom-fields="['Status','Running']"
        :form-custom-fields="['ScannerConfig','Items']"
        @newData="newData"
        @formEditData="editData"
        @formBeforeSubmit="beforeSubmit"
        @cancelEdit="showItemDlg=false"
      > 
        <template v-slot:list_item_Status="item">
          <v-chip small color="green" dark v-if="item.Status=='Active'">{{ item.Status }}</v-chip>
          <v-chip small color="error" dark v-if="item.Status!='Active'">{{ item.Status }}</v-chip>
        </template>

        <template v-slot:list_item_Running="item">
          <template v-if="item.Status=='Active'">
            <v-btn color="warning" x-small v-if="item.Running!='Running'" @click="startPipe(item)">
             <v-icon left>mdi-play</v-icon>Click to Run
            </v-btn>

            <v-btn color="green" x-small v-if="item.Running=='Running'" @click="stopPipe(item)">
             <v-icon left>mdi-stop</v-icon>Click to Stop
            </v-btn>
          </template>
        </template>

        <template v-slot:form_item_ScannerConfig>
          <b>Scanner Config</b>
          <v-textarea
            v-model="ScannerConfigM"
            outlined
            rows="3"
          />
        </template>
        
        <template v-slot:form_item_Items>
          <b>Worker Items</b>
          <v-list dense>
            <v-list-item v-for="pi in workerItemArray" :key="'pipe-item-'+pi.ID"
              style="cursor:pointer"
            >
              <v-list-item-icon @click="deleteItem(pi)">
                <v-btn small icon color="warning"><v-icon>mdi-delete</v-icon></v-btn>
              </v-list-item-icon>
              <v-list-item-content @click="editItem(pi)">
                <v-list-item-title>{{ pi.ID }}, {{ pi.WorkerID }}, {{ pi.Routes.length }} route(s)</v-list-item-title>
                <v-list-item-subtitle style="font-weight:normal;font-size:0.9em">
                  {{  getItemSubtitle(pi) }}
                </v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
            <v-list-item>
              <v-list-item-icon @click="addItem">
                <v-btn small icon color="secondary"><v-icon>mdi-plus-box</v-icon></v-btn>
              </v-list-item-icon>
            </v-list-item>
          </v-list>
        </template>
      </k-browser-2>

      <v-navigation-drawer
        v-if="showItemDlg"
        absolute
        permanent
        right
        :width="600"
        style="z-index:10000"
      >
        <v-toolbar height="20" flat dense class="ma-2">
          <v-spacer />
          <v-icon small color="primary" class="ml-2" @click="showItemDlg=false">mdi-arrow-expand-right</v-icon>
        </v-toolbar>

        <v-card-text>
          <k-form
            ref="frmItem"
            meta="/pipeitem/formconfig"
            :source="pipeItem"
            :mode="itemDlgMode"
            :hide-default-submit="itemDlgMode=='edit'"
            @doSubmit="savePipeItem"
          >
            <template v-slot:item_Config>
              <b>Worker Config</b>
              <v-textarea
                v-model="workerConfigM"
                outlined
                rows="3"
              />
            </template>

            <template v-slot:item_Routes>
              <b>Routes</b>
              <k-grid-2
                ref="gridRoutes"
                meta="/pipeitemroute/gridconfig"
                :use-inline-editor="true"
                :auto-add-line="false"
                :show-select="false"
                :show-search="false"
                :show-reload="false"
                :show-footer="false"
                :source="routes"
                :sourceParm="{itemsPerPage:-1}"
              >
              </k-grid-2>
            </template>

            <template v-slot:buttons>
              <v-btn small color="warning" @click="showItemDlg=false" v-if="itemDlgMode=='new'">
                <v-icon left>mdi-cancel</v-icon>
                Cancel
              </v-btn>
            </template>
          </k-form>
        </v-card-text>

      </v-navigation-drawer>
    </v-card>
  </v-container>
</template>

<script>
import KBrowser2 from '@shared/components/k-vue/KBrowser2.vue'
import KForm from '@shared/components/k-vue/KForm.vue'
import KGrid2 from '@shared/components/k-vue/KGrid2.vue'
import dimension from '@/mixin/dimension.js'

export default {
  components: { KBrowser2, KForm, KGrid2 },
  name: 'DataPipe',
  mixins: [dimension],
  data () {
    return {
      ScannerConfigM: '',
      workerConfigM: '',
      showItemDlg: false,
      itemDlgMode: '',
      pipeItem: {},
      workerItems: {},
      routes: []
    }
  },

  computed: {
    workerItemArray () {
      const res = []
      const keys = Object.keys(this.workerItems)
      for (const keyIndex in keys) {
        res.push(this.workerItems[keys[keyIndex]])
      }
      return res
    }
  },

  watch: {
    showItemDlg (nv) {
      if (this.itemDlgMode=='edit' && nv===false) {
        this.$refs.frmItem.submitForm()
      }
    }
  },

  methods: {
    newData (item) {
      item.Status = 'Inactive'
      this.ScannerConfigM = '{}'
      this.workerItems = {}
    },

    editData (item) {
      this.ScannerConfigM = JSON.stringify(item.ScannerConfig)
      this.workerItems = item.Items
    },

    startPipe (item) {
      this.$axios.post('coordinator/startpipe', item._id).
        then(r => {
          this.$refs.db.refresh()
        }, e => {
          this.$tool.error(e)
        })
    },

    stopPipe (item) {
      this.$axios.post('coordinator/stoppipe', item._id).
        then(r => {
          this.$refs.db.refresh()
        }, e => {
          this.$tool.error(e)
        })
    },

    beforeSubmit (item) {
      if (this.showItemDlg) this.$refs.frmItem.submitForm()
      item.ScannerConfig = JSON.parse(this.ScannerConfigM)
      item.Items = this.workerItems
    },

    addItem () {
      this.showItemDlg = true
      this.itemDlgMode = 'new',
      this.pipeItem = {}
      this.workerConfigM = '{}'
      this.routes = []
    },

    editItem (pi) {
      if (this.showItemDlg && this.itemDlgMode=='edit' && this.pipeItem.ID && this.pipeItem.ID!=pi.ID) {
        this.$refs.frmItem.submitForm()
      }
      
      this.showItemDlg = true
      this.itemDlgMode = 'edit',
      this.pipeItem = pi
      this.routes = pi.Routes ? pi.Routes : []
      this.workerConfigM = JSON.stringify(pi.Config)
    },

    deleteItem (pi) {
      let res = {}
      this.workerItemArray.
        filter(x => x.ID != pi.ID).
        map(x => {
          res[x.ID]=x
        })
      this.workerItems = res
    },

    getItemSubtitle (pi) {
      const res = []
      res.push(pi.CollectProcess ? 'Collect' : 'Distributed')
      if (pi.CloseWhenDone) res.push('Close if Done')
      if (pi.CloseWhenFail) res.push('Close if Fail')
      return res.join(', ')
    },

    savePipeItem (item) {
      let res = {}
      Object.keys(this.workerItems).map(x => {
        res[x] = this.workerItems[x]
      })
      try {
        item.Config = JSON.parse(this.workerConfigM)
      } catch(e) {}
      item.Routes = this.$refs.gridRoutes.dataItems()
      res[item.ID] = item
      this.workerItems = res
      if (this.showItemDlg) this.showItemDlg = false
    }
  }
}
</script>