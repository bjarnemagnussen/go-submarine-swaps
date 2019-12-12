
(function ($) {
    "use strict";


    /*==================================================================
    [ Focus Contact2 ]*/
    $('.input100').each(function(){
        $(this).on('blur', function(){
            if($(this).val().trim() != "") {
                $(this).addClass('has-val');
            }
            else {
                $(this).removeClass('has-val');
            }
        })
    })


    /*==================================================================
    [ Validate ]*/
    var invoice = $('.validate-input input[name="invoice"]');
    var deposit = $('.validate-input select[name="deposit"]');


    $('.validate-form').on('submit',function(){
        var check = true;

        if($(invoice).val().trim() == ''){
            showValidate(invoice);
            check=false;
        }

        if($(deposit).val() == 'Choose Deposit Currency'){
            showValidate(deposit);
            check=false;
        }

        var c = checkValid();
        if(c != ''){
          alertInvalid(c);
          check = false;
        }

        // TODO-MAYBE: Validate Bech32 Lightning string before submitting?
        // if($(email).val().trim().match(/^([a-zA-Z0-9_\-\.]+)@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.)|(([a-zA-Z0-9\-]+\.)+))([a-zA-Z]{1,5}|[0-9]{1,3})(\]?)$/) == null) {
        //     showValidate(email);
        //     check=false;
        // }

        return check;
    });



    $(invoice).change(
      function() {
        var c = checkValid();
        if(c != ''){
          alertInvalid(c)
        } else {
          hideInvalid();
        }
      }
    );


    $(deposit).change(
      function() {
        var c = checkValid();
        if(c != ''){
          alertInvalid(c);
        } else {
          hideInvalid();
        }
      }
    );



    $('.validate-form .input100').each(function(){
        $(this).focus(function(){
           hideValidate(this);
       });
    });


    $('.selection-2').change(
        function() {
          hideValidate(this);
        }
    );



    function checkValid() {
      var result = '';
      if($(deposit).val() != 'Choose Deposit Currency' && $(invoice).val() != ''){
        $.ajax({
          url: 'ajaxvalidateform',
          type: 'post',
          dataType: 'text',
          data : {
            deposit: $(deposit).val(),
            invoice: $(invoice).val()
          },
          async: false,
          success : function(data) {
            result = data;
          },
        });
      };
      return result;
    }



    function showValidate(input) {
        var thisAlert = $(input).parent();

        $(thisAlert).addClass('alert-validate');
    }

    function hideValidate(input) {
        var thisAlert = $(input).parent();

        $(thisAlert).removeClass('alert-validate');
    }



    function alertInvalid(data) {
      $(invoice).parent().attr('data-validate', data);
      showValidate(invoice);
    }

    function hideInvalid() {
      $(invoice).parent().attr('data-validate', 'Lightning invoice is required');
      hideValidate(invoice);
    }



})(jQuery);
