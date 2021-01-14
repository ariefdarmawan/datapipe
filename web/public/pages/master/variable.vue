<template>
  <v-container fluid>
    <v-card outlined>
      <v-card-title>
        Variable
      </v-card-title>

      <k-browser-2
        :list-inline-editor="true"
        :list-show-footer="false"
        :list-auto-add-line="false"
        list-mode='grid'
        list-meta='/variable/gridconfig'
        list-source='/variable/gets'
        list-save-url='/variable/save'
        list-delete-url="/variable/delete"
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
  name: 'MasterVariable',
  data () {
    return {
      objectM: ''
    }
  }
}
</script>