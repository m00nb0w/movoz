Segmented control / view switcher. `tone="dock"` is the dark floating tool bar from the wireframe; mark toggle items `accent` for the terracotta "Annotations" / "Grid" buttons.

```jsx
<Tabs value={v} onChange={setV} items={['Home','Article']} />
<Tabs tone="dock" value={v} onChange={setV}
  items={['Home','Article',{value:'Annotations',label:'Annotations',accent:true},{value:'Grid',label:'Grid',accent:true}]} />
```
