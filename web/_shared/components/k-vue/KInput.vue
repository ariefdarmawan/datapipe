<template>
  <div>
    <v-text-field  
      v-model="inputValue"
      ref="input_tf"
      v-if="fieldType!='bool' && fieldType!='html' && items.length==0 && lookupUrl=='' && multiRow==1 && !multiple && !useList" 
      :label='label'
      :readonly='readOnly'
      :dense="dense"
      :hide-details="hideDetails"
      :single-line='dense==true'
      :type="hideValue ? 'password' : fieldType=='password' ? 'text':fieldType"
      :rules="rules"
      :append-icon="masked || fieldType=='password' ? hideValue ? 'mdi-eye-off' : 'mdi-eye' : ''"
      @focus="onFocus"
      @blur="onBlur"
      @keyup.enter="onKeyEnter"
      @click:append="hideValue=!hideValue"
    />

    <v-textarea 
      v-model="inputValue"
      outlined dense
      ref="input_ta"
      :hide-details="hideDetails"
      :label='label'
      :readonly='readOnly'
      :type="masked?'password':fieldType"
      :rules="rules"
      :rows="multiRow"
      v-if="fieldType!='bool' && fieldType!='html' && items.length==0 && lookupUrl=='' && multiRow>1" 
      @focus="onFocus"
      @blur="onBlur"
    />

    <div v-if="fieldType=='html'">{{ label }}</div>
    <wysiwyg v-model="inputValue" v-if="fieldType=='html'" />

    <template v-if="fieldType!='bool' && !allowAdd && !multiple && useList">
      <v-autocomplete
        v-model="selected"
        v-if='lookupUrl==""'
        :hide-details="hideDetails"
        :items="listItems"
        :readonly='readOnly'
        :dense='dense==true'
        :single-line='dense==true'
        :multiple='multiple'
        :label='label'
        @focus="onFocus"
        @blur="onBlur"
      ></v-autocomplete>

      <v-autocomplete
        v-model="selected"
        v-if='lookupUrl!=""'
        :loading="loading"
        :hide-details="hideDetails"
        :items="listItems"
        :readonly='readOnly'
        :dense='dense==true'
        :single-line='dense==true'
        :multiple='multiple'
        :label='label'
        :search-input.sync='lookup'
        @focus="onFocus"
        @blur="onBlur"
      >
        <template v-slot:no-data>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>
                No results matching "<strong>{{ lookup }}</strong>"
              </v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>
      </v-autocomplete>
    </template>

    <template v-if="fieldType!='bool' && allowAdd && !multiple && useList">
      <v-combobox
        v-model="selected"
        v-if='lookupUrl==""'
        :items="listItems"
        :hide-details="hideDetails"
        :readonly='readOnly'
        :dense='dense==true'
        :single-line='dense==true'
        :multiple='multiple'
        :label='label'
        @focus="onFocus"
        @blur="onBlur"
      ></v-combobox>

      <v-combobox
        v-model="selected"
        v-if='lookupUrl!=""'
        :loading="loading"
        :hide-details="hideDetails"
        :items="listItems"
        :readonly='readOnly'
        :dense='dense==true'
        :single-line='dense==true'
        :multiple='multiple'
        :label='label'
        :search-input.sync='lookup'
        @focus="onFocus"
        @blur="onBlur"
      >
        <template v-slot:no-data>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>
                No results matching "<strong>{{ lookup }}</strong>"
              </v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>
      </v-combobox>
    </template>

    <template v-if="fieldType!='bool' && !allowAdd && multiple && useList">
      <v-autocomplete
        v-model="selecteds"
        v-if='lookupUrl==""'
        :items="listItems"
        :hide-details="hideDetails"
        :readonly='readOnly'
        :dense='dense==true'
        :single-line='dense==true'
        :multiple='multiple'
        :label='label'
        @focus="onFocus"
        @blur="onBlur"
      ></v-autocomplete>

      <v-autocomplete
        v-model="selecteds"
        v-if='lookupUrl!=""'
        :loading="loading"
        :hide-details="hideDetails"
        :items="listItems"
        :readonly='readOnly'
        :dense='dense==true'
        :single-line='dense==true'
        :multiple='multiple'
        :label='label'
        :search-input.sync='lookup'
        @focus="onFocus"
        @blur="onBlur"
      >
        <template v-slot:no-data>
          <v-list-item>
            <v-list-item-content>
              <v-list-item-title>
                No results matching "<strong>{{ lookup }}</strong>"
              </v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>
      </v-autocomplete>
    </template>

    <template v-if="fieldType!='bool' && allowAdd && multiple && useList">
      <v-combobox
        v-model="selecteds"
        v-if="lookupUrl==''"
        :items="listItems"
        :hide-details="hideDetails"
        :readonly='readOnly'
        :dense='dense==true'
        :single-line='dense==true'
        :multiple='multiple'
        :hide-selected='allowAdd'
        :label='label'
        :small-chips='multiple'
        color="secondary"
        @focus="onFocus"
        @blur="onBlur"
      >
        <template v-slot:no-data>
          <v-list-item v-if="allowAdd">
            <v-list-item-content>
              <v-list-item-title>
                No results matching "<strong>{{ lookup }}</strong>". Press <kbd>enter</kbd> to create a new one
              </v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>
      </v-combobox>

      <v-combobox
        v-model="selecteds"
        v-if="lookupUrl!=''"
        :loading="loading"
        :hide-details="hideDetails"
        :items="listItems"
        :readonly='readOnly'
        :dense='dense==true'
        :single-line='dense==true'
        :multiple='multiple'
        :hide-selected='allowAdd'
        :label='label'
        :search-input.sync='lookup'
        small-chips='multiple'
        color="secondary"
        @focus="onFocus"
        @blur="onBlur"
      >
        <template v-slot:no-data>
          <v-list-item v-if="allowAdd">
            <v-list-item-content>
              <v-list-item-title>
                No results matching "<strong>{{ lookup }}</strong>". Press <kbd>enter</kbd> to create a new one
              </v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </template>
      </v-combobox>
    </template>

    <v-checkbox
      v-if="fieldType==='bool'"
      :readonly='readOnly'
      :label="label"
      v-model="boolString"
      @change="onBlur"
      @blur="onBlur"
    ></v-checkbox>
  </div>
</template>

<script>
export default {
  name: 'KInput',

  props: {
    value: { type: [String, Number, Boolean, Array], default: '' },
    kind: { type: String, default: 'text' },
    multiple: { type: Boolean, default: false },
    allowAdd: { type: Boolean, default: false },
    useList: { type: Boolean, default: false },
    hideDetails: { type: Boolean, default: false },
    items: {
      type: Array,
      default () {
        return []
      }
    },
    rules: {
      type: Array,
      default () {
        return []
      }
    },
    multiRow: {type: Number, default: 1},
    reference: { type: String, default: '' },
    masked: { type: Boolean, default: false },
    label: { type: String, default: '' },
    fieldType: { type: String, default: 'string' },
    dense: { type: Boolean, default: false },
    readOnly: { type: Boolean, default: false },
    lookupUrl: { type: String, default: '' },
    lookupKey: { type: String, default: '' },
    lookupFields: {
      type: Array,
      default () {
        return []
      }
    },
    boolValue: {type: Boolean}
  },

  data () {
    return {
      listItems: [],
      hideValue: this.masked || this.fieldType=='password',
      selected: {value:'', text:''},
      selecteds: [],
      inputValue: this.format(this.value),
      loading: false,
      lookup: null,
      inputData: null,
      boolString: this.fieldType=='bool' ? this.value : false
    }
  },

  watch: {
    value (nv) {
      if (!this.useList){
       this.inputValue = this.format(nv)
       return
      }

      if (nv!='' && !this.multiple) {
        const datas = this.listItems.filter(x => {
          if (typeof x=='object') {
            return x.value == nv
          }
          return x == nv
        })

        if (datas.length > 0) {
          this.selected = datas[0]
        } else {
          this.selected = {
            text: nv, value: nv
          }
        }
        return
      } else if (nv!='' && this.multiple) {
        //console.log('selecteds set from value:', JSON.stringify(nv))
        this.selecteds = nv
        return
      }

      if (nv=='') {
        this.selected = {value:'', text:''}
        if (this.selecteds.length > 0) this.selecteds = []
      } else if (this.lookupUrl!='') {
        const filter = {
          op: '$eq',
          field: this.lookupKey,
          value: nv
        }

        this.$axios({
          url: this.lookupUrl,
          method: 'post',
          data: {
            where: filter
          }
        }).then(r => {
            this.listItems = r.data.map(d => {
              if (d[this.lookupKey]==nv) {
                this.selected ={
                  value: d[this.lookupKey],
                  text: d[this.lookupFields[0]]
                }
              }
              return {
                value: d[this.lookupKey],
                text: d[this.lookupFields[0]]
              }
            })
            this.loading = false
          }
        )
      }
    },

    inputValue (nv) {
      if (this.fieldType!='bool') {
        this.$emit('input', this.encode(nv))
      }
    },

    boolString (nv) {
      if (this.fieldType=='bool') {
        this.$emit('input', nv)
      }
    },

    selected (nv) {
      if (!this.useList || this.multiple) return

      if (typeof nv === 'object' && nv!=null) {
        this.inputValue = nv.value
      } else {
        this.inputValue = nv
      }
      //console.log('post value selected: ' + this.inputValue)
    },

    selecteds (nv) {
      if (!this.multiple) return
      this.inputValue = nv
      //console.log('post value selected:', JSON.stringify(this.inputValue))
    },
    
    lookup (val) {
      this.buildListItem(val)
    },

    boolValue: function(val) {
      this.boolString = val;
    }
  },
 
  mounted () {
    if (!this.useList) return

    if (this.lookupUrl!='') {
      this.buildListItem('')
      return
    }
    
    this.listItems = this.items.map(x => {
      let parts = x.split('::')
      if (parts.length==1) {
        return {
          value: parts[0],
          text: parts[0]
        }
      } else {
        return {
          value: parts[0],
          text: parts[1]
        }
      }
    })
    
    if (!this.multiple) {
      const sels = this.listItems.filter(x => x.value==this.value)
      if (sels.length > 0) {
        this.selected = sels[0]
      } else {
        this.selected = {text:this.value, value:this.value}
      }
    } else {
      this.selecteds = this.value
    }
  },

  methods: {
    focus () {
      if (this.$refs.input_tf) {
        this.$refs.input_tf.focus()
        return
      }

      if (this.$refs.input_ta) {
        this.$refs.input_ta.focus()
        return
      }
    },

    buildListItem (val) {
      //console.log('val:',val,'selected:',JSON.stringify(this.selected))
      if (val==undefined || val==null || (val==this.selected.value && val!='')) return
      //console.log('val:',val,'selected:',JSON.stringify(this.selected))
      if (this.loading) return

      let addKeyQuery = true
      if (this.lookupUrl==''){
        return
      }
      
      this.listItems = []
      this.loading = true
      const filters = this.lookupFields.map(l => {
        if (l === this.lookupKey) {
          addKeyQuery = false
        }
        return {
          op: '$contains',
          field: l,
          value: [val]
        }
      })
      if (addKeyQuery) {
        filters.push({
          op: '$contains',
          field: this.lookupKey,
          value: [val]
        })
      }
      const filter =
        filters.length === 1 ? filters[0] : {
          op: '$or',
          items: filters
        }
      this.$axios({
        url: this.lookupUrl,
        method: 'post',
        data: {
          where: filter,
          take: 30
        }
      }).then(r => {
          this.listItems = r.data.map(d => {
            const itemText = this.lookupFields.map(x => d[x]).filter(x => x!='').join(" ")

            const item =  {
              value: d[this.lookupKey],
              text: d[this.lookupFields[0]]
            }

            if (this.useList && !this.multiple && item.value==this.value) {
              this.selected = item
            }

            return item
          })
          this.loading = false
        }
      )
    },

    onFocus() {
      if (this.fieldType!='bool') {
        this.inputData = this.inputValue
      } else {
        //console.log('value: ' + this.boolString)
        this.inputData = this.boolString
      }
    },

    onKeyEnter () {
      this.$emit('keyEnter', !this.boolString, this.boolString, this.reference)
    },

    onBlur() {
      if (this.fieldType!='bool')
        this.$emit('blur', this.inputData, this.inputValue, this.reference)
      else
        this.$emit('blur', !this.boolString, this.boolString, this.reference)
    },

    format (value) {
      let res = value
      switch (this.fieldType) {
        case 'date':
          res = this.$moment(value).format('yyyy-MM-DD')
          break;

        case 'time':
          if (value==undefined || value==null){
            value=0
          }
          let hr = Math.floor(value / 100)
          let mn = value % 100
          res = hr < 10 ? '0' + hr.toString() : hr.toString()
          res += ':'     
          res += mn < 10 ? '0' + mn.toString() : mn.toString()
      }
      return res
    },
    
    encode (value) {
      let res = value
      switch (this.fieldType) {
        case 'time':
          let parts = value.split(":")
          let hr = parseInt(parts[0])
          let mn = parseInt(parts[1])
          res = hr * 100 + mn
          break

        case 'numeric':
          res = parseFloat(res)
      } 
      return res
    }
  }
}
</script>

<style>
.v-input--selection-controls {
  margin-top: 0px;
  padding-top: 4px;
}

.v-textarea textarea{
  line-height: 1.2em;
}
</style>