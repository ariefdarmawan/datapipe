<template>
  <div>
    <template v-if="!metaIsLoaded">
      ... please wait while building form ...
    </template>

    <template v-else>
      <div>
        <v-tabs 
          :height="35"
          v-model="tabModel" 
          v-show="showTabs"
          :background-color="tabBgColor"
          :slider-color="tabSliderColor"
          :slider-size="4"
          @change="tabChange"
          dark
        >
          <v-tab 
            v-for="(tab, tabidx) in cTabs"
            :key="'tab-'+tabidx">{{ tab }}</v-tab>
          </v-tabs>

          <v-tabs-items v-model="tabModel">
            <v-tab-item 
              v-for="(tab, tabidx) in cTabs"
              :key="'tab-'+tabidx"
            >
              <div v-show="tabidx == 0">
                <v-form
                  :loading="loading"
                  v-model="formIsValid"
                  ref="form"
                  v-show="!loading && metaIsLoaded && ((hideAfterSubmitOK && !submitted) || !hideAfterSubmitOK)"
                >
                  <div v-for="(grp,gidx) in formConfig.fieldgroups" :key="'form-group-s'+gidx">
                    <div v-if="grp.ShowTitle" class="kf-group-title">{{ grp.Title }}</div>
                    <v-row v-for="(row,idx) in formConfig.rows[grp.Title]" :key="'form-row-'+gidx+'-'+idx" no-gutters>
                      <v-col v-for="(col, colidx) in row.columns" :key="'form-row-'+idx+'-col-'+colidx" class="mr-4">
                        <slot :name="'item_'+col.field" v-bind="{item,col}">
                          <div v-if="col.fieldType == 'list-editor'">
                            <b>{{ col.label }}</b>
                            <KListEditor
                              v-model="item[col.field]"
                              :meta="col.SubListConfig"
                              :form-meta="col.SubFormConfig"
                              :show-search="false"
                              :show-navigation="false"
                            />
                          </div>

                          <div v-else-if="col.fieldType == 'bool'">
                            <KInput
                              v-model="item[col.field]"
                              :boolValue="item[col.field]"
                              :masked="col.masked"
                              :readOnly="col.readOnly || readOnly"
                              :label="col.label"
                              :dense="dense"
                              :hide-details="!showInputMessage"
                              :rules="fieldRules(col)"
                              :fieldType="col.fieldType"
                              :items="col.listItems ? col.listItems : []"
                              :lookupUrl="col.lookupUrl"
                              :lookupKey="col.lookupKey"
                              :lookupFields="col.lookupFields"
                            />
                          </div>

                          <div v-else>
                            <KInput
                              v-model="item[col.field]"
                              :masked="col.masked"
                              :readOnly="col.readOnly || (col.isKey && mode=='edit') || readOnly"
                              :label="col.label"
                              :dense="dense"
                              :hide-details="!showInputMessage"
                              :rules="fieldRules(col)"
                              :fieldType="col.fieldType"
                              :multiRow="col.multiRow"
                              :useList="col.useList"
                              :allowAdd="col.allowAdd"
                              :multiple="col.multiple"
                              :items="col.listItems && col.listItems.length > 0 ? col.listItems : []"
                              :lookupUrl="col.lookupUrl"
                              :lookupKey="col.lookupKey"
                              :lookupFields="col.lookupFields"
                            />
                          </div>
                        </slot>
                      </v-col>
                    </v-row>
                  </div>
                  <slot name="append-form" v-bind:item="item" />
                  <v-row>
                    <v-col class="mx-auto">
                      <v-btn color="primary" @click="submitForm" small v-show='!hideDefaultSubmit' :width="submitFill ? '100%' : ''">
                        <v-icon left>mdi-content-save</v-icon>
                        {{ submitText }}
                      </v-btn>
                      <slot name="buttons" />
                    </v-col>
                    <v-col align="right" v-if="showClose" @click="$emit('close')">
                      <v-btn icon small>
                        <v-icon>mdi-close</v-icon>
                      </v-btn>
                    </v-col>
                  </v-row>
                  <slot name="footer" v-bind:item="item" />
                </v-form>
              </div>

              <div v-if="tabidx > 0" class="mt-5">
                <slot :name="'tab-'+tabidx" v-bind="item">
                </slot>
              </div>
            </v-tab-item>
          </v-tabs-items>  
          <br/>
      </div>
    </template>
  </div>
</template>

<script>
import KInput from './KInput.vue'

export default {
  name: 'KForm',

  components: {
    KInput, 
    KListEditor: () => import('./KListEditor.vue')
  },

  props: {
    value: {type: Object, default: () => null},
    mode: {type: String, default:'edit'},
    meta: {type: [String, Object], default:''},
    readAfterMeta: { type: Boolean, default: true },
    source: { type: [String, Function, Object], default:'' },
    sourceParm: { type: [String, Function,Object,Array], default:'' },
    post: {type: String, default:''},
    tabs: { type: Array, default: () => []},
    ignoreAxiosError: { type: Boolean, default: false },
    showInputMessage: { type: Boolean, default: true },
    dense: { type: Boolean, default: false },
    showTabs: { type: Boolean, default: false },
    submitText: { type: String, default: 'Save' },
    showClose: { type: Boolean, default: false },
    submitFill: { type: Boolean, default: false },
    hideAfterSubmitOK: { type: Boolean },
    hideDefaultSubmit: { type: Boolean, default: false },
    tabBgColor: { type: String, default: 'secondary' },
    tabSliderColor: { type: String, default: 'yellow' },
  
    successText: { type: String, default: 'Success' },
    readOnly: { type: Boolean, default: false },
  },

  data () {
    return {
      tabModel: null,
      loading: false,
      metaIsLoaded: false,
      item: {},
      formConfig: null,
      formIsValid: true,
      dirty: false,
      submitted: false,
      errorTxt: '',
      rules: {
        required: value => !!value || 'Required',
        charCounter(value, min, max) {
          // var value=this.item?this.item[name]?this.item[name]:"":""
          if (max === 0) {
            max = 10000
          }

          const l = value.length
          if (l < min || l > max) {
            return 'entry length should be between ' + min + ' and ' + max
          }

          return true
        }
      }
    }
  },

  watch: {
    meta (nv) {
      this.loadMeta()
    },

    source (nv) {
      this.readFromSource()
    },

    sourceParm (nv) {
      this.readFromSource()
    }
  },

  computed: {
    cTabs () {
      const newTabs = this.tabs.map(x=>{return x})
      if (newTabs.length==0) {
        newTabs.push(this.formConfig.Title ? this.formConfig.Title : 'General')
      }
      newTabs.push('Back')
      return this.mode=='new' ? [newTabs[0]] : newTabs
    }
  },

  mounted () {
    this.loadMeta()
  },

  methods: {
    loadMeta() {
      const tof = typeof this.meta

      this.metaIsLoaded = false
      if (tof === 'function') {
        this.formConfig = this.meta()
        this.metaIsLoaded = true
        if (this.readAfterMeta) this.readFromSource()
      } else if (tof === 'string' && this.meta !== '') {
        this.loading = true
        this.formConfig = {}
        this.$axios({
          url: this.meta,
          method: 'post',
          data: ''
        }).then(
          r => {
            this.loading = false
            this.formConfig = r.data
            this.metaIsLoaded = true

            // add additional
            this.formConfig.fieldgroups.forEach(g => {
              this.formConfig.rows[g.Title].forEach(r => {
                r.columns.forEach(c => {
                  if (this.formConfig.keyfields) {
                    this.formConfig.keyfields.forEach(k => {
                      if (k === c.field) {
                        c.isKey = true
                      } else {
                        c.isKey = false
                      }
                    })
                  }
                })
              })
            })

            if (this.readAfterMeta) this.readFromSource()
          },
          e => {
            this.loading = false
            this.$tool.error(e)
          }
        )
      } else if (tof === 'object') {
        this.formConfig = this.meta
        this.metaIsLoaded = true
        if (this.readAfterMeta) this.readFromSource()
      }
    },

    fieldRules(item) {
      const self = this
      const rules = []

      if (item.required) {
        rules.push(this.rules.required)
      }

      if (item.fieldType === 'numeric') {
        rules.push(function(value) {
          if (isNaN(value)) {
            return 'value should be a numeric'
          }
          return true
        })
      }

      if (item.minLength > 0 || item.maxLength > 0) {
        rules.push(function(value) {
          return self.rules.charCounter(value, item.minLength, item.maxLength)
        })
      }

      if (item.minValue || item.maxValue) {
        rules.push(function(value) {
          if (item.minValue && value < item.minValue) {
            return (
              'value should be between ' +
              item.minValue +
              ' and ' +
              item.maxValue
            )
          }

          if (item.maxValue && value > item.maxValue) {
            return (
              'value should be between ' +
              item.minValue +
              ' and ' +
              item.maxValue
            )
          }

          return true
        })
      }

      return rules
    },

    validate() {
      if (this.item) {
        let valRes = this.$emit('validate', this.item)
        if (valRes==undefined || valRes){
          return this.$refs.form[0].validate()
        }
      }
      return false
    },

    setLoading(p) {
      this.loading = p
    },

    setSubmitted(p) {
      this.submitted = p
    },

    setErrorTxt(p) {
      this.errorTxt = p
    },


    readFromSource () {
      const parm = top=='function' ? this.sourceParm() : this.sourceParm     
      const tos = typeof this.source
      switch (tos) {
        case 'string':
          if (this.source=='') {
            this.item = {}
            return
          }

          this.$axios.post(this.source, parm).
            then(r => {
              console.log('post result: '+JSON.stringify(r.data))
              this.item = r.data
            }, e => {
              this.item = {}
              this.$emit('assignDefault', this.item)
              if (this.ignoreAxiosError)this.$tool.error(e)
            })
          break

        case 'object':
          this.item = this.source
          this.$emit('assignDefault', this.item)
          this.tabModel = 0
          //console.log('form-source-object: '+JSON.stringify(this.item))
          break

        case 'function':
          const top = typeof this.sourceParm
          this.item = this.source(parm)
          this.$emit('assignDefault', this.item)
          break
      }
    },

    tabChange (tabNo) {
      this.$emit('tabChange', tabNo)
    },

    formData () {
      const parm = this.item
      this.formConfig.fieldgroups.forEach(g => {
        this.formConfig.rows[g.Title].forEach(r => {
          r.columns.forEach(c => {
            // handle date
            const dt = parm[c.field]
            switch (c.fieldType) {
              case 'date':
                if (parm[c.field]) {
                  parm[c.field] = this.$moment(dt).format('YYYY-MM-DDT00:00:00Z')
                }
                break
              
              case 'number', 'numeric':
                if (!isNaN(dt)) {
                  parm[c.field] = parseFloat(dt)
                } else {
                  parm[c.field] = 0
                }
                break
            }
          })
        })
      })

      return parm
    },

    submitForm() {
      //console.log("vaidate process")
      this.errorTxt = ''
      if (!this.validate()) {
        //console.log("invaidated")
        return
      }
      //console.log("validated")

      let parm = this.formData()
      
      this.loading = true
      this.$emit('beforeSubmit', parm)
      if (this.post === '') {
        this.$emit('doSubmit', parm)
        this.$emit('afterSubmit', parm)
        this.submitted = true
        this.loading = false  
        return
      }
      
      //console.log("data to be saved:",JSON.stringify(parm))
      this.$axios.post(this.post,parm).then(
        r => {
          this.submitted = true
          this.loading = false
          if (this.showSnackbar) {
            this.snackbar = true
            this.snackbarText = this.successText
            this.$tool.info(this.successText)
          }
          this.$emit('afterSubmit', r.data)
        },
        e => {
          let txt = e
          this.$tool.error(e)
          this.loading = false
        }
      )
    }
  }
}
</script>