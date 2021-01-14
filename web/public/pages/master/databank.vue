<template>
  <v-container fluid>
    <v-card outlined>
      <v-card-title>
        Data Bank
      </v-card-title>

      <k-browser-2
        :list-inline-editor="true"
        :list-show-footer="true"
        :list-auto-add-line="false"
        list-mode='grid'
        list-meta='/databank/gridconfig'
        list-source='/databank/gets'
        list-save-url='/databank/save'
        list-delete-url="/databank/delete"
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
  name: 'MasterDataBank',
  data () {
    return {
    }
  }
}
</script>