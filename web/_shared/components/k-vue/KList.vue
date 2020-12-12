<template>
  <div>
    <v-row>
      <v-col v-if="showSearch" cols="9" xs="12">
        <v-text-field 
          single-line
          dense
          placeholder="Enter search keyword"
          v-model="keyword"
          @change="changeSearch"
          @click:append="refresh"
          append-icon="mdi-refresh"
        />
      </v-col>
      <v-col>
        <v-btn small text @click="newItem" v-if="showNew" color="primary">
          <v-icon small left>mdi-plus</v-icon>New
        </v-btn>
        <v-btn small text color="warning"
        :disabled="!selected || selected.length==0" @click="deleteItems" v-if="showDelete">
          <v-icon small left>mdi-delete</v-icon>Delete
        </v-btn>
      </v-col>
    </v-row>

    <div v-if="loading">
      ... please wait while loading ...
    </div>

    <v-list
      dense
      v-if="!loading"
      :two-line="viewLine===2"
      :three-line="viewLine===3"
    >
      <template v-for="(item,idx) in items">
        <v-list-item :key="'list-item-'+idx">
          <v-list-item-action v-if="showSelect">
            <v-checkbox
              :value="item[metaCfg.keyField]"
              v-model="selected"
            ></v-checkbox>
          </v-list-item-action>
          <v-list-item-content>
            <slot name="content" v-bind="{data: item, index: idx}">
              <v-list-item-title 
                @click="selectItem(item)" 
                style="cursor:pointer">{{ getItemValue(item, metaCfg.titleFields) }}</v-list-item-title>
              <v-list-item-subtitle 
                @click="selectItem(item)" 
                v-for="(stFields, stidx) in metaCfg.subtitleFields" :key="'st-'+stidx"
                style="font-weight:normal"
              >
                {{ getItemValue(item, stFields) }}
              </v-list-item-subtitle>
            </slot>
          </v-list-item-content>
          <v-list-item-icon v-if="!hideAction">
            <slot name="action" v-bind="{data: item, index: idx}">
              <v-btn icon>
                <v-icon color="primary lighten-1" @click="selectItem(item)">mdi-information</v-icon>
              </v-btn>
            </slot>
          </v-list-item-icon>
        </v-list-item>
        <v-divider 
          :inset="showSelect"
          v-if="idx < (items.length - 1)"
          :key="'divider-'+idx"></v-divider>
      </template>
    </v-list>

    <v-row v-if="showNavigation">
      <v-col class="right">
        <span>Showing {{ items.length }} of {{ dataCount }}, page {{ currentPage }}&nbsp;</span>
        <v-btn
          class="ma-2"
          small text
          :disabled="options.skip==0"
          @click="prev"
        >
          <v-icon left>mdi-arrow-left</v-icon>
          Prev
        </v-btn>
        <v-btn
          class="ma-2"
          small text
          :disabled="upperBound >= dataCount"
          @click="next"
        >
          Next
          <v-icon right>mdi-arrow-right</v-icon>
        </v-btn>
      </v-col>
    </v-row>

    <v-row class="mb-5" v-if="showNewBottom">
      <v-col>
        <v-btn small text @click="newItem" color="primary">
          <v-icon small left>mdi-plus</v-icon>New
        </v-btn>
      </v-col>
    </v-row>
  </div>
</template>

<script>
export default {
  name: "KList",

  props: {
    value: { type: Array, default: () => {
      return []
    }}, 
    showSelect: { type: Boolean, default: true },
    debug: { type: Boolean, default: false },
    source: { type: [String, Function, Array] , default: '' },
    meta: { type: [String, Object] , default: '' },
    sourceParm: { type: Object, default: () => {return {
      itemsPerPage: 10
    }}},
    viewLine: { type: Number, defaut: 1},
    refreshOnMount: {type: Boolean, default: true},
    hideFooter: { type: Boolean, default: false },
    hideAction: { type: Boolean, default: false },
    showSearch: { type: Boolean, default: true },
    showEdit: { type: Boolean, default: true },
    showNavigation: { type:Boolean, default: true },
    showNew: { type: Boolean, default: true },
    showNewBottom: { type: Boolean, default: false },
    showDelete: { type: Boolean, default: true },
    showRefresh: { type: Boolean, default: true },
    readOnly: { type: Boolean, default: false },
  },

  watch: {
    value (nv) {
      this.items = nv
    },

    items (nv) {
      this.$emit('input', nv)
    },

    source (nv) {
      this.refresh()
    },

    sourceParm (nv) {
      this.refresh()
    }
  },

  computed: {
    upperBound () {
      return this.options.skip + this.items.length
    },

    currentPage () {
      const cp = (this.options.skip / this.options.itemsPerPage) + 1
      return isNaN(cp) ? 1 : cp
    },

    pageSize () {
      return this.sourceParm.itemsPerPage ? this.sourceParm.itemsPerPage : 10
    }
  },

  data () {
    return {
      keyword: '',
      options: {
        take: this.sourceParm.itemsPerPage,
        skip: 0
      },
      metaCfg: {},
      dataCount: 0,
      loading: false,
      actionWidth: '50px',
      searchFields: [],
      headers: [],
      selected: [],
      currentSortField: '',
      sortDescending: false,
      items: [],
      infoLog: '',
      metaLoaded: false
    }
  },

  mounted () {
    this.loadMeta()
    if (this.refreshOnMount) {
      this.refresh()
    }
  },

  methods: {
    newItem() {
      this.$emit('create')
    },

    selectItem(item) {
      this.$emit('select', item)
    },

    deleteItems() {
      if (this.metaCfg.keyField === '') {
        return
      }
      if (
        !confirm(
          'You are about to delete ' +
            this.selected.length +
            ' record(s). Are you sure ?'
        )
      ) {
        return
      }
      this.$emit('delete', this.selected.map(x => {
        return [x]
      }))
      this.selected = []
    },

    loadMeta() {
      if (typeof this.meta=='object') {
        this.metaCfg = this.meta
        return
      }

      if (typeof this.meta=='string' && this.meta=='') return
      this.loading = true
      this.$axios.post(this.meta, {}).then(
        r => {
          this.loading = false
          this.metaCfg = r.data
          this.infoLog = r
          this.metaLoaded = true
        },
        e => {
          this.loading = false
        }
      )
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

    next() {
      if (!this.options.skip) this.options.skip = 0
      this.options.skip += this.pageSize
      this.refresh()
    },

    prev() {
      this.options.skip -= this.pageSize
      this.refresh()
    },

    changeSearch () {
      this.options.skip = 0
      this.refreshAPI()
    },

    refresh () {
      const tos = typeof this.source
      let parm = typeof this.sourceParm=='function' ? this.sourceParm() : this.sourceParm
      //console.log(tos)
      
      switch (tos) {
        case 'function':
          this.items = this.source(parm)
          break

        case 'object':
          //console.log(tos)
          this.items = this.source
          break

        case 'string':
          this.refreshAPI()
          break
      }
    },

    refreshAPI () {
      const parm = this.options
      parm.where = null
      
      // keyword
      let wheres = []
      if (this.keyword != '') {
        const searchFields =
          this.metaCfg.searchFields.length > 0 ? this.metaCfg.searchFields : ['_id', 'Name']
        const searchOps = searchFields.map(it => {
          return { op: '$contains', field: it, value: [this.keyword] }
        })
        if (searchOps.length==1) wheres.push(searchOps[0])
          else if (searchOps.length > 1) wheres.push({ op: '$or', items: searchOps })
      }
      if (this.sourceParm && this.sourceParm.where && this.sourceParm.where!=null) {
        wheres.push(this.sourceParm.where)
      }
      if (wheres.length==1) {
        parm.where = wheres[0]
      } else if (wheres.length > 1){
        parm.where = {op: '$and', items: wheres}
      }

      // sort
      if (this.currentSortField !== '') {
        parm.sort = [
          this.sortDescending
            ? '-' + this.currentSortField
            : this.currentSortField
        ]
      }

      this.loading = true
      this.$axios.post(this.source, parm).then(
        r => {
          this.selected = []
          let items = []
          // if (r.data !== '') {
          this.dataCount = r.data.count
          items = r.data.data.map(d => {
            this.headers.forEach(c => {
              // handle date
              if (c.type === 'date' && d[c.value]) {
                const dt = d[c.value]
                d[c.value] = this.$moment(dt).format(c.format)
              }
            })
            return d
          })

          this.items = items
          this.loading = false
        },

        e => {
          this.$tool.error(e)
          this.loading = false
        }
      )
    },
  }
}
</script>