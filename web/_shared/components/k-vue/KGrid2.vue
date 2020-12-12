<template>
  <div>
    <v-toolbar dense flat color="primary" height="32" dark>
      <v-text-field 
        v-model="keyword"
        hide-details single-line dense 
        height="34"
        placeholder="keyword pencarian"></v-text-field>
      <v-spacer></v-spacer>
      <v-btn small @click="reload" text>
        <v-icon left>mdi-refresh</v-icon>
        Reload
      </v-btn>
      <v-divider vertical />
      <v-btn small text 
        v-if="showNew"
        @click="newData" :disabled='!allItemHasBeenSynced || editIndex > -1'>
        <v-icon left>mdi-plus-box</v-icon>
        New
      </v-btn>
    </v-toolbar>
    <v-data-table
      dense fixed-header
      :height="height"
      :item-key="config.gridKey"
      :show-select="showSelect"
      :items="items"
      :items-per-page="sourceParm.itemsPerPage ? sourceParm.itemsPerPage : -1"
      :loading="loading"
      :headers="tableHeaders"
      :hide-default-footer="!showFooter"
      @dblclick:row="dtEditRow"
    >
      <template v-for="(col, idx) in tableHeaders" v-slot:[col.slotName]="{item}">
          <div :key="'k-item-data-'+idx" v-if="col.field != 'k_options'">
            <div v-show="item.mode && item.mode=='edit' && !col.readOnly">
              <slot :name="'item_' + col.field + '_edit'" v-bind="item">
                <KInput 
                  v-model="item[col.field]"
                  :dense="true"
                  :required="col.required"
                  :readOnly="col.readOnly || (col.isKey && entryMode=='edit')"
                  :reference="item[config.gridKey] + '|' + col.field"  
                  :key="'input_' + idx"
                  :ref="'input_' + idx"
                  :field-type="col.fieldType"
                  :useList="col.useList"
                  :items="col.listItems"
                />
              </slot>
            </div>

            <div v-show="!(item.mode && item.mode=='edit')">
              <slot :name="'item_' + col.field + ''" v-bind="item">
                <template v-if="col.fieldType!='bool'">{{ item[col.field] }}</template>
                <v-icon v-if="col.fieldType=='bool' && item[col.field]" color="green" small>mdi-check</v-icon>
              </slot>
            </div>
          </div>

          <div :key="'k-item-opts-'+idx" v-if="col.field == 'k_options'">
            <slot name="buttons" v-bind="item">
              <v-btn 
                icon x-small color="primary"
                v-if="item.mode=='edit'"
                @click="submitLine"><v-icon>mdi-content-save</v-icon></v-btn>

              <v-btn 
                icon x-small color="warning"
                v-if="item.mode=='edit'"
                @click="editIndex=-1"><v-icon>mdi-cancel</v-icon></v-btn>

              <v-btn 
                icon x-small color="primary"
                v-if="editIndex==-1"
                @click="edit(item)"><v-icon>mdi-border-color</v-icon></v-btn>

              <v-btn 
                icon x-small color="warning"
                v-if="editIndex==-1 && showDelete"
                @click="deleteItem(item)"><v-icon>mdi-delete</v-icon></v-btn>
            </slot>
          </div>
      </template>
    </v-data-table>

    <div id="debug" v-if="showDebug" class="ma-4">
      props: {{ $props }}<br/><br/>
      config: {{ config }}<br/><br/>
      editindex: {{ editIndex }} <br/>
      {{ items }}
    </div>
  </div>
    
</template>

<script>
import Vue from 'vue'
import KInput from './KInput.vue'

export default {
  name: 'KGrid2',

  components: {
    KInput
  },

  props: {
    value: { type:Object, default: ()=>{}},
    meta: { type: String, default: '' },
    source: { type: [String, Function, Object], default: '' },
    sourceParm: { type: Object, default: ()=>{}},
    gridType: { type: String, default: 'simple' },
    height: { type: [String, Function, Object], default: '' },
    showDebug: { type: Boolean, default: false },
    showSelect: { type: Boolean, default: true },
    showDelete: { type: Boolean, default: true },
    showFooter: { type: Boolean, default: true },
    showOptions: { type: Boolean, default: true },
    showNew: { type: Boolean, default: true},
    inlineEditor: { type: Boolean, default: true},
    autoAddLine: { type: Boolean, default: true},
    externalEditor: { type: Boolean, default: false},
    saveUrl: { type: String, default: ''},
    deleteUrl: { type: String, default: ''}
  },

  data() {
    return {
      config: {
        gridKey: 'GridKey',
        headers: []
      },
      editEventSource: '',
      keyword: '',
      dataCount: 0,
      loading: false,
      editIndex: -1,
      entryMode: '',
      items: [],
    }
  },

  mounted () {
    this.loadMeta()
    this.reload()
  },

  watch: {
    editIndex (nv) {
      let items = this.items.map(x => {
        if (x.GridKey=='') this.assignGridKeyToItem(x)
        x.mode = 'view'
        return x
      })
      this.items = items

      if (nv >= 0 && nv < items.length) {
        this.$emit('input', this.items[nv])
      } else {
        this.$emit('input', {})
      }

      if (nv >= 0 && nv < items.length) {
        const useInlineEditor = this.inlineEditor && 
          (this.editEventSource=='dblClick' || !this.externalEditor)
        if (useInlineEditor) {
          this.items[this.editIndex].mode = 'edit'
          Vue.nextTick(() => {
            if (this.$refs.input_0 && this.$refs.input_0[this.editIndex])
              this.$refs.input_0[this.editIndex].focus()
          })
        } else {
          this.$emit('editData', this.items[this.editIndex])
        }
      }
    }
  },

  computed: {
    tableHeaders () {
      return this.config.headers.filter(x => x.show=='show')
    },

    itemInEdit () {
      return this.editIndex < 0 ? {} : 
        this.editIndex >= this.items.length ? {} :
          this.items[editIndex]
    },

    allItemHasBeenSynced () {
      const unsync = this.items.filter(x => x.sync!==true)
      return unsync.length == 0
    }
  },

  methods: {
    setEditIndex (i) {
      this.editIndex = i
    },

    newData () {
      if (!this.externalEditor) {
        this.addLine()
        return
      }

      this.$emit('newData')
    },

    dataItems () {
      return this.items
    },

    componentConfig () {
      return this.config
    },

    dtEditRow (ev, obj) {
      const editedItems = this.items.filter(x => x.mode=='edit')
      if (editedItems.length > 0) return
      this.edit(obj.item, 'dblClick')
    },

    edit (item, source) {
      let index = -1
      let i = -1

      this.items.forEach(d => {
        i++
        if (d.GridKey==item.GridKey) {
          //console.log(i, JSON.stringify(d), JSON.stringify(item))
          index = i
        }
      })

      this.editEventSource = source ? source : 'click'
      this.entryMode = 'edit'
      if (index>=0) this.editIndex = index
    },

    isAllKeyExist(item) {
      let valid = false
      this.config.headers.forEach(x => {
        if (x.isKey && item[x.field] && item[x.field]!='' ) valid = true
      })
      return valid
    },

    deleteItem (item) {
      if (this.deleteUrl=='') {
        if (this.$listeners.deleteData) {
          this.$emit('deleteData', item, ()=>{
            this.items = this.items.filter(x => x.GridKey != item.GridKey)
          })
          return
        }

        this.items = this.items.filter(x => x.GridKey != item.GridKey)
        return
      }

      if (!this.isAllKeyExist(item)) {
        this.items = this.items.filter(x => x.GridKey != item.GridKey)
        return
      }

      this.$axios.post(this.deleteUrl, item).
        then(r => {
          this.items = this.items.filter(x => x.GridKey != item.GridKey)
        }, e => {
          this.$tool.error(e)
        })
    },

    assignGridKeyToItem (item) {
      item.GridKey='key-' + Math.random().toString(36).replace(/[^a-z]+/g, '') 
    },

    assignGridKeyfromItemKey (item, fields) {
      const keyValues = fields.
        filter(x => x.IsKey).
        sort((a,b) => a==b ? 0 : a<b ? -1 : 1).
        map(x => item[x.field])
      item.GridKey = keyValues.join('-')
    },

    copyData (source, dest, fields) {
      fields.forEach(x => {
        dest[x.field] = source[x.field]
      })
    },

    submitLine () {
      if (!this.validate()) {
        this.$tool.error('ada error di form')
        return false
      }

      if (this.saveUrl == '' ) {
        this.items[this.editIndex].sync = true
        if (this.autoAddLine && this.editIndex==this.items.length-1) {
          this.addLine()
        } else {
          this.editIndex = -1
        }
      } else {
        let parmPost = this.items[this.editIndex]
        if (this.transformData!=undefined) {
          parmPost = this.transformData(parmPost)
        }
        this.$axios.post(this.saveUrl, parmPost).then(r => {
          let item = {}
          this.copyData(r.data, item, this.config.headers)
          this.assignGridKeyfromItemKey(item, this.config.headers)
          item.sync = true
          this.items[this.editIndex] = item
          if (this.autoAddLine && this.editIndex==this.items.length-1) this.addLine()
            else this.editIndex = -1
        }, e => {
          this.$tool.error(e)
        })
      }

      return true
    },

    validate () {
      if (this.editIndex<0) return false
      //if (this.items[this.editIndex].Name=='' || this.items[this.editIndex].Name==undefined) return false
      return true
    },
 
    addLine () {
      this.editIndex = -1
      
      const item = {}
      this.assignGridKeyToItem(item)
      item.mode = 'view'
      this.items.push(item)
      
      Vue.nextTick(() => {
        this.entryMode = 'new'
        this.editIndex = this.items.length - 1
      })
    },

    loadMeta () {
      this.loading = true

      switch (typeof this.meta) {
        case 'object':
          this.config = this.meta
          this.loading = false
          return

        case 'function':
          this.config = this.meta()
          this.loading = false
          return

        case 'string':
          if (this.meta=='') {
            this.config = {
              headers: [],
              gridKey: 'GridKey'
            }
            this.loading = false
          }

          this.$axios.post(this.meta, {}).then(
            r => {
              //this.config.gridKey = r.data.keyField
              const headers = 
                r.data.fields.
                map(function(x) {
                  let data = {
                    show: x.show,
                    field: x.field,
                    text: x.label,
                    align: x.fieldType === 'number' ? 'right' : x.align,
                    fieldType: x.fieldType,
                    required: x.required,
                    readOnly: x.readOnly,
                    isKey: x.isKey,
                    value: x.field,
                    slotName: 'item.' + x.field,
                    format: x.format,
                    useList: x.useList,
                    listItems: x.listItems,
                    lookupUrl: x.lookupUrl,
                    lookupKey: x.lookupKey,
                    lookupFields: x.lookupFields
                  }

                  if (x.fieldType=='bool') {
                    data.width = 120
                  }

                  return data
                })

                if (this.showOptions) 
                  headers.push({
                    field: 'k_options',
                    show: 'show',
                    value: 'k_options',
                    text: 'Options',
                    slotName: 'item.k_options',
                    fieldType: 'string',
                    width: 120
                  })

                this.config.headers = headers
                this.config.keyField = r.data.keyField
                this.config.searchFields = r.data.searchFields
                this.loading = false
            },
            e => {
              this.loading = false
            }
          )
      }
    },

    reload () {
      if (typeof this.source=='string' && this.source!='') {
        this.loading = true

        const parm = {}
        for (const attr in this.sourceParm) {
          parm[attr] = this.sourceParm[attr]
        }

        // build a where
        if (this.keyword && this.keyword!='') {
          const filters = this.config.searchFields.map(x => {
            return {
              field: x, op: '$contains', value: [this.keyword]
            }
          })
          parm.where = {op: '$or', items: filters}
        }

        this.$axios.post(this.source, parm).
          then(r => {
            const items = r.data.data.map(x => {
              this.assignGridKeyToItem(x)
              x.sync = true
              return x
            })
            this.items = items
            this.dataCount = r.data.count
            this.loading = false
            if (this.autoAddLine && !this.externalEditor) this.addLine()
          }, e => {
            this.$tool.error(e)
            this.loading = false
            if (this.autoAddLine && !this.externalEditor) this.addLine()
          })
        return
      }

      if (this.autoAddLine) this.addLine()
    },

    // have to make this function for backward compat
    refresh () {
      this.reload()
    }
  }
}
</script>