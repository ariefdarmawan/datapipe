<template>
  <div>
    <k-grid-2
      v-show="browserMode=='list'"
      ref="grid"
      v-model="selected"
      :meta="listMeta"
      :source="listSource"
      :source-parm="listSourceParm"
      :auto-add-line="listAutoAddLine"
      :inline-editor="listInlineEditor"
      :save-url="listSaveUrl"
      :delete-url="listDeleteUrl"
      :new-url="listNewUrl"
      :show-new="showNew"
      :show-search="showSearch"
      :show-reload="showReload"
      :show-select="showSelect"
      :show-footer="listShowFooter"
      :show-delete="showDelete"
      :external-editor="formSave!=''"
      :prevent-dbl-click="preventDblClick"
      @editData="editData"
      @newData="newData"
    >
      <template v-for="nm in listCustomSlotNames" v-slot:[nm]="item">
        <slot :name="'list_'+nm" v-bind="item">
        </slot>
      </template>
      <template v-slot:buttons="item">
        <slot name="list_buttons" v-bind="item">
        </slot>
      </template>
      <template v-slot:extra_buttons="item">
        <slot name="list_extra_buttons" v-bind="item">
        </slot>
      </template>
    </k-grid-2>
    
    <k-form 
      v-if="browserMode=='form'"
      class="ma-2"
      :meta="formMeta"
      :mode="formMode"
      :source="formDataSource"
      :post="formSave"
      :tabs="formTabs"
      :showTabs="formTabs.length>1"
      @beforeSubmit="beforeSubmit"
      @doSubmit="formSubmit"
      @afterSubmit="formAfterSubmit"
    >
      <template v-slot:buttons>
        <slot name="form-buttons-1">
          <v-btn color="warning" small @click="cancelEdit">
            <v-icon left>mdi-cancel</v-icon>
            Cancel
          </v-btn>
        </slot>
        <slot name="form-buttons-2">
        </slot>
      </template>

      <template v-for="nm in formCustomSlotNames" v-slot:[nm]="item">
        <slot :name="'form_'+nm" v-bind="item">
        </slot>
      </template>

      <template v-for="tabName in formTabSlotNames" v-slot:[tabName]="item">
        <slot :name="'form_'+tabName" v-bind="item">
        </slot>
      </template>
    </k-form>
  </div>
</template>

<script>
import KForm from './KForm.vue'
import KGrid2 from './KGrid2.vue'
export default {
  components: { KGrid2, KForm },
  name: 'KBrowser2',

  props: {
    mode: {type:String, default:'list'},
    showSelect: {type:Boolean, default:false},
    showSearch: {type:Boolean, default:true},
    showReload: {type:Boolean, default:true},
    showNew: {type:Boolean, default:true},
    showDelete: {type:Boolean, default:true},
    preventDblClick: {type:Boolean, default:false},
    listAutoAddLine: {type:Boolean, default:false},
    listInlineEditor: {type:Boolean, default:false},
    listShowFooter: {type:Boolean, default:true},
    listSaveUrl: {type:String, default:''},
    listDeleteUrl: {type:String, default:''},
    listNewUrl: {type:String, default:''},
    listMode: {type:String, default:'grid'},
    listMeta: {type: [String, Function, Object], default:''},
    listSource: {type: [String, Function, Object], default:''},
    listSourceParm: {type: Object, default: () => {
      return {
        itemsPerPage: 10
      }
    }},
    listCustomFields: {type: Array, default: () => []},
    formMeta: {type: [String, Function, Object], default:''},
    formSource: {type: [String, Function, Object], default: () => {}},
    formSourceParm: {type: [String, Function, Object], default:''},
    formCustomFields: {type: Array, default: () => []},
    formTabs: {type: Array, default: () => []},
    formSave: {type: String, default:''},
  },

  data () {
    return {
      selected: {},
      formDataSource: {},
      browserMode: this.mode,
      formMode: 'edit'
    }
  },

  computed: {
    listCustomSlotNames () {
      return this.listCustomFields.map(x=>{
        return 'item_' + x
      })
    },

    formCustomSlotNames () {
      return this.formCustomFields.map(x=>{
        return 'item_' + x
      })
    },

    formTabSlotNames () {
      if (this.formTabs.length<2){
        return []
      }

      const res = []
      this.formTabs.forEach((tab,idx)=>{
        res.push('tab-'+idx)
      })
      return res
    },

    listConfig () {
      return this.listMode == 'grid' ? this.$refs.grid.componentConfig() : this.$refs.list.componentConfig()
    },

    listComponent () {
      return this.listMode == 'grid' ? this.$refs.grid : this.$refs.list
    }
  },

  methods: {
    refresh () {
      this.browserMode="list"
      this.$refs.grid.refresh()
    },

    cancelEdit () {
      this.$emit('cancelEdit')
      this.listComponent.setEditIndex(-1)
      this.browserMode = 'list'
    },

    editData (item) {
      this.browserMode='form'
      this.formMode='edit'
      const config = this.listConfig
      const id = [item[config.keyField]]
      this.$axios.post(this.formSource,id).
        then(r => {
          const data = r.data
          this.$emit('formEditData', data)
          this.formDataSource = data
        }, e => {
          this.$tool.error(e);
        })
    },

    newData () {
      this.browserMode='form'
      this.formMode='new'
      const newRecord = {}
      this.$emit("newData", newRecord)
      this.formDataSource = newRecord
    },

    beforeSubmit (item) {
      this.$emit('formBeforeSubmit', item)
    },

    formSubmit (item) {
      this.$emit('formSubmit', item)
    },

    formAfterSubmit (item) {
      this.listComponent.setEditIndex(-1)
      this.browserMode = 'list'
      this.listComponent.refresh()
      this.$emit('formAfterSubmit', item)
    }
  }
}
</script>