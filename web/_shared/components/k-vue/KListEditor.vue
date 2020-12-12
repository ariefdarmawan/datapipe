<template>
  <div>
    <template v-if="viewMode=='list'">
      <KList
        :meta="listMeta"
        :source="items"
        :show-search="showSearch"
        :show-navigation="showNavigation"
        :show-select="false"
        :show-delete="false"
        :show-new="false"
        :show-new-bottom="editIndex < 0"
        @create="createItem"
      >
        <template v-slot:content="item">
          <KListForm
            ref="formAnswerOption"
            v-if="item.data.viewMode=='edit'"
            :show-input-message="false"
            :dense="true"
            :meta="formMeta"
            :source="item.data"
            :hide-default-submit="true"
            @doSubmit="onOptionSubmission"
          />
          <template v-if="item.data.viewMode!='edit'">
            <v-list-item-title @click="edit(item.index)"
              style="cursor:pointer">{{ getItemValue(item.data, listMeta.titleFields) }}</v-list-item-title>
            <v-list-item-subtitle 
              v-for="(stFields, stidx) in listMeta.subtitleFields" :key="'st-'+stidx"
              style="font-weight:normal"
            >
              {{ getItemValue(item.data, stFields) }}
            </v-list-item-subtitle>
          </template>
        </template>
        <template v-slot:action="item">
          <v-btn x-small icon 
            v-if="item.data.viewMode=='edit'"
            @click="saveAnswerOption">
            <v-icon>mdi-content-save</v-icon>
          </v-btn>
          <v-menu
            v-if="item.data.viewMode!='edit'"
            x-small
            offset-y
          >
            <template v-slot:activator="{ attrs, on }">
              <v-btn icon x-small
                v-bind="attrs"
                v-on="on"
              >
                <v-icon>mdi-menu</v-icon>
              </v-btn>
            </template>
            <v-list min-width="100">
              <v-list-item @click="edit(item.index)" :disabled="editIndex >= 0">Edit</v-list-item>
              <v-list-item @click="rmItem(item.index)" :disabled="editIndex >= 0">Delete</v-list-item>
            </v-list>
          </v-menu>
        </template>
      </KList>
    </template>
  </div>
</template>

<script>
import KList from './KList.vue'

export default {
  name: 'KListEditor',

  components: {
    KListForm: () => import('./KForm.vue'),
    KList
  },

  props: {
    value: { type: Array, default: () => {
      return []
    }},
    showSelect: { type: Boolean, default: true },
    source: { type: [String, Function, Object] , default: '' },
    meta: { type: [String, Object] , default: '' },
    formMeta: { type: [String, Object] , default: '' },
    sourceParm: { type: Object, default: () => {return {
      itemsPerPage: 10
    }}},
    dataItem: { type: Array },
    viewLine: { type: Number, defaut: 1},
    refreshOnMount: {type: Boolean, default: true},
    hideFooter: { type: Boolean, default: false },
    hideAction: { type: Boolean, default: false },
    showSearch: { type: Boolean, default: true },
    showEdit: { type: Boolean, default: true },
    showNavigation: { type:Boolean, default: true },
    showNew: { type: Boolean, default: true },
    showDelete: { type: Boolean, default: true },
    showRefresh: { type: Boolean, default: true },
    readOnly: { type: Boolean, default: false },
  },

  data () {
    return {
      viewMode: 'list',
      listMeta: '',
      editMode: false,
      editIndex: -1,
      items: this.value
    }
  },

  watch: {
    value (nv) {
      this.items = nv
      this.editMode = false
    },

    source (nv) {
      this.readDataFromSource()
    },

    sourceParm (nv) {
      this.readDataFromSource()
    },

    items (nv) {
      this.$emit("input",nv)
    }
  },

  mounted () {
    this.loadMeta()
  },

  methods: {
    loadMeta () {
      if (typeof this.meta=='object' && this.meta) {
        this.listMeta = this.meta
      }

      if (typeof this.meta=='string' && this.meta!='') {
        this.$axios.push(this.meta).then(r => {
          this.listMeta = r.data
          //this.readDataFromSource()
        }, e => {
          this.$tool.error(e)
        })
      }
    }, 

    readDataFromSource () {
      this.items = []
        
      if (typeof this.source=='object') {
        if (typeof this.sourceParm=='function') {
          this.items = this.sourceParm(this.source)
        }
        this.items = this.source
        return
      }

      if (this.source=='') return

      let parm = {}
      if (typeof this.sourceParm=='object') {
        parm = this.sourceParm
      } else if (typeof this.sourceParm=='function') {
        parm = this.sourceParm()
      }

      this.$axios.post(this.source, parm).
        then(r => {
          this.items = r.data
        }, e => {
          this.$tool.error(e)
        })
    },

    edit (index) {
      if (this.editIndex >= 0) {
        this.items[this.editIndex].viewMode = 'list'
      }

      this.editMode = true
      this.editIndex = index
      this.items[index].viewMode = 'edit'
    },

    getItemValue (item, fields) {
      if (fields==undefined || fields.length==0) return
      const arr = fields.map(x => {
        return item[x]
      }).filter(x => {
        return x && x!=null && x!=''
      })
      return arr.join(' | ')
    },

    rmItem (index) {
      const newItems = this.items.
        filter((x, idx) => {
          return idx!=index
        })
      //console.log(newItems)
      this.items = newItems
    },

    saveAnswerOption () {
      this.$refs.formAnswerOption.submitForm()
    },

    cancelAnswerOption () {
      this.items[this.editIndex].viewMode = ''
      this.editMode = false
      this.editIndex = -1
    },

    onOptionSubmission (item) {
      item.viewMode='list' 
      let newItems = this.items.map((x,i) => {
        return i==this.editIndex ? item : x
      })
      this.items = newItems
      this.editMode = false
      this.editIndex = -1
    },

    createItem () {
      const data = {}
      this.$emit('assignDefault', data)
      if (!this.items) this.items = []
      data.viewMode = 'edit'
      this.editMode = true
      this.editIndex = this.items.length
      this.items.push(data)
      return this.items
   }
  }
}
</script>