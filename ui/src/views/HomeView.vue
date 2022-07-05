<script>
import {onMounter} from "vue"
import {reactive} from "vue"
import {useRouter} from "vue-router"

export default {
  setup() {
    onMounter(async () => {
      const url = "http://localhost:5000/user"
      const res = await fetch(url, {
        credentials: true
      })
      if (res.status !== 200) {
        console.log(res.statusText)
      }
      console.log(res.data)

    })
    const data = reactive({
      email: "",
      user: "",
    })
    const router = useRouter()
    return {data, router}
  }
}
</script>

<template>
  <main>
    <h1>Hello, {{ data.user }}</h1>
  </main>
</template>
