<template>
  <v-container fluid>
    <v-card outlined>
      <v-card-title>
        Storage
      </v-card-title>

      <k-browser-2
        list-mode='grid'
        list-meta='/storage/gridconfig'
        list-source='/storage/gets'
        form-meta='/storage/formconfig'
        form-source='/storage/get'
        form-save='/storage/save'
        list-delete-url="/storage/delete"
        :form-custom-fields="['Data']"
        @newData="newData"
        @formEditData="editData"
        @formBeforeSubmit="beforeSubmit"
      > 
        <template v-slot:form_item_Data>
          <b>Data</b>
          <v-textarea
              v-model="objectM"
              outlined
              rows="5"
          />
        </template>
      </k-browser-2>
    </v-card>
  </v-container>
</template>

<script>
import KBrowser2 from '@shared/components/k-vue/KBrowser2.vue'
export default {
  components: { KBrowser2 },
  name: 'MasterStorage',
  data () {
    return {
      objectM: ''
    }
  },

  methods: {
    newData (item) {
      this.objectM = '{}'
    },

    editData (item) {
      this.objectM = JSON.stringify(item.Data)
    },

    beforeSubmit (item) {
      item.Data = JSON.parse(this.objectM)
    }
  }
}
</script>